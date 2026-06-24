package transformer

import (
	"go/ast"
	"go/token"
	"strings"
	"testing"

	"github.com/gomphermp/gomphermp/internal/parser"
)

// newCanonicalForStmt builds a detached but structurally valid
// `for i := 0; i < N; i++ {}` node. It passes extractLoopVar and
// extractUpperBound, so it is useful for exercising downstream branches like
// the replaceForStmt failure path without tripping the earlier guards.
func newCanonicalForStmt() *ast.ForStmt {
	return &ast.ForStmt{
		Init: &ast.AssignStmt{
			Lhs: []ast.Expr{&ast.Ident{Name: "i"}},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "0"}},
		},
		Cond: &ast.BinaryExpr{
			X:  &ast.Ident{Name: "i"},
			Op: token.LSS,
			Y:  &ast.Ident{Name: "N"},
		},
		Post: &ast.IncDecStmt{
			X:   &ast.Ident{Name: "i"},
			Tok: token.INC,
		},
		Body: &ast.BlockStmt{},
	}
}

// TestTransform_For_BasicRewrite verifies the canonical for transformation:
// //gompher for over a loop becomes runtime.For(func(i int) { body }, N),
// preserving the loop variable name and extracting the upper bound.
func TestTransform_For_BasicRewrite(t *testing.T) {
	src := `package main

func main() {
	const N = 10
	data := make([]int, N)
	//gompher for
	for i := 0; i < N; i++ {
		data[i] = i * i
	}
	_ = data
}
`
	got := runTransform(t, src)

	if !strings.Contains(got, `runtime.For(threadID, func(i int) {`) {
		t.Errorf("expected runtime.For(threadID, func(i int) {...}, N), got:\n%s", got)
	}
	if !strings.Contains(got, "data[i] = i * i") {
		t.Errorf("expected loop body preserved, got:\n%s", got)
	}
	if !strings.Contains(got, "}, N)") {
		t.Errorf("expected bound N passed as second argument, got:\n%s", got)
	}
}

// TestTransform_ParallelFor_BasicRewrite verifies the combined construct:
// //gompher parallel for becomes runtime.ParallelFor(func(i int) { body }, N).
func TestTransform_ParallelFor_BasicRewrite(t *testing.T) {
	src := `package main

func main() {
	const N = 16
	results := make([]int, N)
	//gompher parallel for
	for i := 0; i < N; i++ {
		results[i] = i * i
	}
	_ = results
}
`
	got := runTransform(t, src)

	if !strings.Contains(got, `runtime.ParallelFor(func(i int) {`) {
		t.Errorf("expected runtime.ParallelFor(func(i int) {...}, N), got:\n%s", got)
	}
	if !strings.Contains(got, "results[i] = i * i") {
		t.Errorf("expected loop body preserved, got:\n%s", got)
	}
}

// TestTransform_ParallelFor_ScheduleDynamic verifies that a schedule(dynamic,
// c) clause redirects the combined construct to runtime.ParallelForDynamic
// with the requested chunk size, rather than the static ParallelFor entry
// point.
func TestTransform_ParallelFor_ScheduleDynamic(t *testing.T) {
	src := `package main

func main() {
	const N = 20
	results := make([]int, N)
	//gompher parallel for schedule(dynamic, 4)
	for i := 0; i < N; i++ {
		results[i] = heavy(i)
	}
	_ = results
}

func heavy(i int) int { return i }
`
	got := runTransform(t, src)

	if !strings.Contains(got, `runtime.ParallelForDynamic(func(i int) {`) {
		t.Errorf("expected runtime.ParallelForDynamic for parallel for schedule(dynamic), got:\n%s", got)
	}
	if !strings.Contains(got, "}, N, 4)") {
		t.Errorf("expected bound N and chunk 4 as trailing args, got:\n%s", got)
	}
}

// TestTransform_For_ScheduleDynamicDefaultChunk verifies that a dynamic
// schedule with no explicit chunk size defaults to a chunk of 1, matching
// the runtime's own clamping behavior.
func TestTransform_For_ScheduleDynamicDefaultChunk(t *testing.T) {
	src := `package main

func main() {
	const N = 8
	//gompher for schedule(dynamic)
	for i := 0; i < N; i++ {
		work(i)
	}
}

func work(i int) {}
`
	got := runTransform(t, src)

	if !strings.Contains(got, "}, N, 1)") {
		t.Errorf("expected default chunk size 1 for schedule(dynamic) with no chunk, got:\n%s", got)
	}
}

// TestTransform_For_ScheduleStaticUsesStaticFunc verifies that an explicit
// schedule(static) does NOT divert to ForDynamic - static scheduling keeps
// the runtime.For entry point.
func TestTransform_For_ScheduleStaticUsesStaticFunc(t *testing.T) {
	src := `package main

func main() {
	const N = 10
	//gompher for schedule(static)
	for i := 0; i < N; i++ {
		work(i)
	}
}

func work(i int) {}
`
	got := runTransform(t, src)

	if !strings.Contains(got, `runtime.For(threadID, func(i int) {`) {
		t.Errorf("expected runtime.For for schedule(static), got:\n%s", got)
	}
	if strings.Contains(got, "ForDynamic") {
		t.Errorf("expected ForDynamic NOT emitted for static schedule, got:\n%s", got)
	}
	if strings.Contains(got, "StaticChunked") {
		t.Errorf("expected StaticChunked NOT emitted for chunk-less static, got:\n%s", got)
	}
}

// TestTransform_ParallelFor_ScheduleStaticChunked verifies that schedule(static,
// chunk) with an explicit chunk diverts the combined construct to the block-
// cyclic runtime.ParallelForStaticChunked with the requested chunk size.
func TestTransform_ParallelFor_ScheduleStaticChunked(t *testing.T) {
	src := `package main

func main() {
	const N = 20
	//gompher parallel for schedule(static, 5)
	for i := 0; i < N; i++ {
		work(i)
	}
}

func work(i int) {}
`
	got := runTransform(t, src)
	if !strings.Contains(got, `runtime.ParallelForStaticChunked(func(i int) {`) {
		t.Errorf("expected runtime.ParallelForStaticChunked, got:\n%s", got)
	}
	if !strings.Contains(got, "}, N, 5)") {
		t.Errorf("expected bound N and chunk 5 as trailing args, got:\n%s", got)
	}
}

// TestTransform_For_ScheduleStaticChunkedWorksharing verifies that a bare for
// with schedule(static, chunk) uses the worksharing ForStaticChunked (threadID
// first), since the team is provided by the enclosing parallel.
func TestTransform_For_ScheduleStaticChunkedWorksharing(t *testing.T) {
	src := `package main

func main() {
	const N = 12
	//gompher parallel
	{
		//gompher for schedule(static, 3)
		for i := 0; i < N; i++ {
			work(i)
		}
	}
}

func work(i int) {}
`
	got := runTransform(t, src)
	if !strings.Contains(got, `runtime.ForStaticChunked(threadID, func(i int) {`) {
		t.Errorf("expected worksharing ForStaticChunked, got:\n%s", got)
	}
	if !strings.Contains(got, "}, N, 3)") {
		t.Errorf("expected bound N and chunk 3 as trailing args, got:\n%s", got)
	}
}

// TestTransform_For_PreservesLoopVarName verifies that a non-default loop
// variable name (here "idx") is carried through to the closure parameter, so
// the body's references continue to resolve.
func TestTransform_For_PreservesLoopVarName(t *testing.T) {
	src := `package main

func main() {
	const N = 5
	//gompher for
	for idx := 0; idx < N; idx++ {
		consume(idx)
	}
}

func consume(x int) {}
`
	got := runTransform(t, src)

	if !strings.Contains(got, `runtime.For(threadID, func(idx int) {`) {
		t.Errorf("expected closure param named idx, got:\n%s", got)
	}
	if !strings.Contains(got, "consume(idx)") {
		t.Errorf("expected body reference to idx preserved, got:\n%s", got)
	}
}

// TestTransform_For_AddsRuntimeImport verifies the runtime import is injected
// for a for directive in a file that did not previously import it.
func TestTransform_For_AddsRuntimeImport(t *testing.T) {
	src := `package main

func main() {
	const N = 4
	//gompher for
	for i := 0; i < N; i++ {
		work(i)
	}
}

func work(i int) {}
`
	got := runTransform(t, src)

	if !strings.Contains(got, `"github.com/gomphermp/gomphermp/pkg/runtime"`) {
		t.Errorf("expected runtime import in output, got:\n%s", got)
	}
}

// TestTransform_ParallelFor_Reduction verifies the reduction rewrite using the
// pointer-capture technique: capture &sum before the region, shadow sum with a
// per-goroutine accumulator initialised to the operator identity, leave the
// body untouched, and fold each partial back through the pointer under Critical.
func TestTransform_ParallelFor_Reduction(t *testing.T) {
	src := `package main

func main() {
	const N = 100
	sum := 0
	//gompher parallel for reduction(+:sum)
	for i := 0; i < N; i++ {
		sum += i
	}
	_ = sum
}
`
	got := runTransform(t, src)

	for _, want := range []string{
		"_red_sum := &sum",                  // pointer capture before the region
		"runtime.Parallel(func(threadID int) {", // per-goroutine scope
		"var sum int = 0",                   // accumulator with the + identity
		"sum += i",                          // body left untouched
		`runtime.Critical("", func() {`,     // combine under mutual exclusion
		"*_red_sum += sum",                  // fold partial through the pointer
	} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in reduction output, got:\n%s", want, got)
		}
	}
}

// TestTransform_ParallelFor_ReductionMul verifies the multiplicative reduction
// uses identity 1 and the *= combine.
func TestTransform_ParallelFor_ReductionMul(t *testing.T) {
	src := `package main

func main() {
	const N = 10
	prod := 1
	//gompher parallel for reduction(*:prod)
	for i := 0; i < N; i++ {
		prod *= (i + 1)
	}
	_ = prod
}
`
	got := runTransform(t, src)
	if !strings.Contains(got, "var prod int = 1") {
		t.Errorf("expected accumulator with identity 1, got:\n%s", got)
	}
	if !strings.Contains(got, "*_red_prod *= prod") {
		t.Errorf("expected *= combine, got:\n%s", got)
	}
}

// TestTransform_ParallelFor_ReductionMinMax verifies the max/min reductions:
// the accumulator starts from the captured initial value (_init_x) and the
// combine is a conditional through the pointer.
func TestTransform_ParallelFor_ReductionMinMax(t *testing.T) {
	for _, tc := range []struct {
		op  string
		cmp string
	}{
		{"max", "best > *_red_best"},
		{"min", "best < *_red_best"},
	} {
		src := `package main

func main() {
	const N = 10
	best := 0
	//gompher parallel for reduction(` + tc.op + `:best)
	for i := 0; i < N; i++ {
		if i > best {
			best = i
		}
	}
	_ = best
}
`
		got := runTransform(t, src)
		if !strings.Contains(got, "_init_best := best") {
			t.Errorf("%s: expected captured initial value, got:\n%s", tc.op, got)
		}
		if !strings.Contains(got, "var best int = _init_best") {
			t.Errorf("%s: expected accumulator seeded from the initial value, got:\n%s", tc.op, got)
		}
		if !strings.Contains(got, "if "+tc.cmp) {
			t.Errorf("%s: expected conditional combine %q, got:\n%s", tc.op, tc.cmp, got)
		}
	}
}

// TestTransform_ParallelFor_ReductionBoolean verifies the && / || reductions:
// the accumulator identity is true / false and the combine is a logical
// assignment through the pointer.
func TestTransform_ParallelFor_ReductionBoolean(t *testing.T) {
	for _, tc := range []struct {
		op       string
		identity string
		combine  string
	}{
		{"&&", "var ok bool = true", "*_red_ok = *_red_ok && ok"},
		{"||", "var ok bool = false", "*_red_ok = *_red_ok || ok"},
	} {
		src := `package main

func main() {
	const N = 4
	ok := true
	//gompher parallel for reduction(` + tc.op + `:ok)
	for i := 0; i < N; i++ {
		ok = ok && (i >= 0)
	}
	_ = ok
}
`
		got := runTransform(t, src)
		if !strings.Contains(got, tc.identity) {
			t.Errorf("%s: expected identity %q, got:\n%s", tc.op, tc.identity, got)
		}
		if !strings.Contains(got, tc.combine) {
			t.Errorf("%s: expected combine %q, got:\n%s", tc.op, tc.combine, got)
		}
	}
}

// TestReductionVars_InvalidOperator verifies the defensive guard: an operator
// the parser would never emit is rejected. Constructed directly since the
// grammar only produces valid operators.
func TestReductionVars_InvalidOperator(t *testing.T) {
	parsed, err := parser.Parse("package main\n\nfunc main() { x := 0; _ = x }\n")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	_, err = reductionVars(parsed, []parser.Clause{
		parser.ReductionClause{Operator: "/", Vars: []string{"x"}},
	}, parsed.File.End())
	if err == nil || !strings.Contains(err.Error(), "not supported") {
		t.Errorf("expected unsupported-operator error, got: %v", err)
	}
}

// TestTransform_ParallelFor_ReductionSub verifies the subtractive reduction
// uses identity 0 and the -= combine.
func TestTransform_ParallelFor_ReductionSub(t *testing.T) {
	src := `package main

func main() {
	const N = 10
	acc := 0
	//gompher parallel for reduction(-:acc)
	for i := 0; i < N; i++ {
		acc -= i
	}
	_ = acc
}
`
	got := runTransform(t, src)
	if !strings.Contains(got, "var acc int = 0") {
		t.Errorf("expected accumulator with identity 0, got:\n%s", got)
	}
	if !strings.Contains(got, "*_red_acc -= acc") {
		t.Errorf("expected -= combine, got:\n%s", got)
	}
}

// TestTransform_ParallelFor_ReductionDynamic verifies that reduction combines
// with a dynamic schedule: the inner loop is ForDynamic, wrapped by Parallel,
// with the reduction setup intact.
func TestTransform_ParallelFor_ReductionDynamic(t *testing.T) {
	src := `package main

func main() {
	const N = 100
	sum := 0
	//gompher parallel for schedule(dynamic, 4) reduction(+:sum)
	for i := 0; i < N; i++ {
		sum += i
	}
	_ = sum
}
`
	got := runTransform(t, src)
	if !strings.Contains(got, "runtime.ForDynamic(func(i int) {") {
		t.Errorf("expected ForDynamic inside the reduction wrapper, got:\n%s", got)
	}
	if !strings.Contains(got, "}, N, 4)") {
		t.Errorf("expected dynamic chunk 4, got:\n%s", got)
	}
	if !strings.Contains(got, "*_red_sum += sum") {
		t.Errorf("expected reduction combine, got:\n%s", got)
	}
}

// TestTransform_ParallelFor_ReductionUndeclared verifies that a reduction over
// a variable whose type cannot be resolved yields a descriptive error.
func TestTransform_ParallelFor_ReductionUndeclared(t *testing.T) {
	parsed, err := parser.Parse("package main\n\nfunc main() {}\n")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	parsed.Nodes = append(parsed.Nodes, parser.AnnotatedNode{
		Directive: parser.ParallelForDirective{
			Clauses: []parser.Clause{parser.ReductionClause{Operator: "+", Vars: []string{"ghost"}}},
			Node:    newCanonicalForStmt(),
		},
	})
	_, err = Transform(parsed)
	if err == nil || !strings.Contains(err.Error(), "reduction(ghost)") {
		t.Errorf("expected reduction resolution error, got: %v", err)
	}
}

// TestTransform_ParallelFor_Lastprivate verifies the lastprivate rewrite:
// the variable is captured by pointer before the region, declared private per
// goroutine, and the sequentially-last iteration writes its value back.
func TestTransform_ParallelFor_Lastprivate(t *testing.T) {
	src := `package main

func main() {
	const N = 10
	x := 0
	//gompher parallel for lastprivate(x)
	for i := 0; i < N; i++ {
		x = i * 2
	}
	_ = x
}
`
	got := runTransform(t, src)
	for _, want := range []string{
		"_lp_x := &x",                           // pointer capture before the region
		"runtime.Parallel(func(threadID int) {", // per-goroutine scope
		"var x int",                             // private copy per goroutine
		"if i == N-1 {",                         // last-iteration guard
		"*_lp_x = x",                            // write the private value back
	} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in lastprivate output, got:\n%s", want, got)
		}
	}
	// The capture must precede the Parallel wrapper.
	capIdx := strings.Index(got, "_lp_x := &x")
	parIdx := strings.Index(got, "runtime.Parallel")
	if !(capIdx >= 0 && capIdx < parIdx) {
		t.Errorf("expected capture before Parallel; cap=%d par=%d\n%s", capIdx, parIdx, got)
	}
}

// TestTransform_For_Lastprivate verifies lastprivate on a bare worksharing for:
// the capture and private copy live in the per-goroutine block, the writeback
// in the For closure.
func TestTransform_For_Lastprivate(t *testing.T) {
	src := `package main

func main() {
	const N = 8
	//gompher parallel
	{
		last := 0
		//gompher for lastprivate(last)
		for i := 0; i < N; i++ {
			last = i * 3
		}
		_ = last
	}
}
`
	got := runTransform(t, src)
	for _, want := range []string{
		"_lp_last := &last",
		"var last int",
		"runtime.For(threadID, func(i int) {",
		"if i == N-1 {",
		"*_lp_last = last",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in bare-for lastprivate output, got:\n%s", want, got)
		}
	}
}

// TestTransform_ParallelFor_LastprivateLiteralBound verifies the bound is cloned
// for the comparison: a literal upper bound becomes `i == 10-1` cleanly, without
// sharing the node used by the For call.
func TestTransform_ParallelFor_LastprivateLiteralBound(t *testing.T) {
	src := `package main

func main() {
	x := 0
	//gompher parallel for lastprivate(x)
	for i := 0; i < 10; i++ {
		x = i
	}
	_ = x
}
`
	got := runTransform(t, src)
	if !strings.Contains(got, "if i == 10-1 {") {
		t.Errorf("expected literal-bound guard `if i == 10-1`, got:\n%s", got)
	}
}

// TestTransform_ParallelFor_LastprivateUndeclared verifies a lastprivate on an
// undeclared variable surfaces a type-resolution error.
func TestTransform_ParallelFor_LastprivateUndeclared(t *testing.T) {
	parsed, err := parser.Parse("package main\n\nfunc main() {}\n")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	parsed.Nodes = append(parsed.Nodes, parser.AnnotatedNode{
		Directive: parser.ParallelForDirective{
			Clauses: []parser.Clause{parser.LastPrivateClause{Vars: []string{"ghost"}}},
			Node:    newCanonicalForStmt(),
		},
	})
	_, err = Transform(parsed)
	if err == nil || !strings.Contains(err.Error(), "lastprivate(ghost)") {
		t.Errorf("expected lastprivate resolution error, got: %v", err)
	}
}

// TestTransform_ParallelFor_Private verifies that private(x) on a parallel for
// declares a per-goroutine copy inside the Parallel closure (not per iteration).
func TestTransform_ParallelFor_Private(t *testing.T) {
	src := `package main

func main() {
	const N = 8
	tmp := 0
	//gompher parallel for private(tmp)
	for i := 0; i < N; i++ {
		tmp = i
		_ = tmp
	}
	_ = tmp
}
`
	got := runTransform(t, src)
	if !strings.Contains(got, "runtime.Parallel(func(threadID int) {") {
		t.Fatalf("expected Parallel wrapper, got:\n%s", got)
	}
	// The private declaration must sit in the Parallel closure, before the For.
	parIdx := strings.Index(got, "runtime.Parallel")
	declIdx := strings.Index(got, "var tmp int")
	forIdx := strings.Index(got, "runtime.For(threadID")
	if !(parIdx < declIdx && declIdx < forIdx) {
		t.Errorf("expected private decl between Parallel and For; par=%d decl=%d for=%d\n%s", parIdx, declIdx, forIdx, got)
	}
}

// TestTransform_For_Private verifies that private(x) on a bare for wraps the
// worksharing call in a block whose `var x T` is per-goroutine (the block runs
// once per goroutine inside the enclosing parallel).
func TestTransform_For_Private(t *testing.T) {
	src := `package main

func main() {
	const N = 8
	//gompher parallel
	{
		tmp := 0
		//gompher for private(tmp)
		for i := 0; i < N; i++ {
			tmp = i
			_ = tmp
		}
		_ = tmp
	}
}
`
	got := runTransform(t, src)
	if !strings.Contains(got, "var tmp int") {
		t.Errorf("expected private `var tmp int`, got:\n%s", got)
	}
	if !strings.Contains(got, "runtime.For(threadID, func(i int) {") {
		t.Errorf("expected worksharing For inside the block, got:\n%s", got)
	}
}

// TestTransformFor_WrongNodeType verifies the defensive error path when the
// directive's Node is not a *ast.ForStmt.
func TestTransformFor_WrongNodeType(t *testing.T) {
	parsed, err := parser.Parse("package main\n")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	bogus := parser.ForDirective{
		Node: &ast.BlockStmt{},
	}

	err = transformFor(parsed, bogus)
	if err == nil {
		t.Fatal("expected error for non-ForStmt Node")
	}
	if !strings.Contains(err.Error(), "expected *ast.ForStmt") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestTransformFor_NonCanonicalLoop verifies that a loop missing the
// canonical init (for i := 0; ...) is rejected with a descriptive error
// rather than producing malformed output. The extraction helpers require the
// i := 0; i < N; i++ shape.
func TestTransformFor_NonCanonicalLoop(t *testing.T) {
	src := `package main

func main() {
	i := 0
	//gompher for
	for ; i < 10; i++ {
		work(i)
	}
}

func work(i int) {}
`
	parsed, err := parser.Parse(src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	_, err = Transform(parsed)
	if err == nil {
		t.Fatal("expected error for non-canonical loop (missing init)")
	}
	if !strings.Contains(err.Error(), "init") {
		t.Errorf("expected error mentioning the missing init, got: %v", err)
	}
}

// TestTransformFor_InclusiveCondition verifies that an inclusive bound (<=) on a
// bare worksharing for is accepted and normalized: the body iterates over
// [0, N+1) via the synthesized counter, recovering i, so the original range
// [0, N] is preserved.
func TestTransformFor_InclusiveCondition(t *testing.T) {
	src := `package main

func main() {
	const N = 10
	//gompher parallel
	{
		//gompher for
		for i := 0; i <= N; i++ {
			work(i)
		}
	}
}

func work(i int) {}
`
	got := runTransform(t, src)
	dense := strings.ReplaceAll(got, " ", "")
	if !strings.Contains(dense, "i:=_gompherIter") {
		t.Errorf("expected induction recovery `i := _gompherIter`, got:\n%s", got)
	}
	if !strings.Contains(dense, "N-0+1") {
		t.Errorf("expected inclusive trip count `N - 0 + 1`, got:\n%s", got)
	}
}

// TestTransformFor_ForStmtNotInAST verifies that a structurally valid loop
// that is nevertheless not reachable from the file AST produces a "for
// statement not found" error rather than silently dropping the runtime call.
// This is the loop analog of the BodyNotInAST guard for block directives.
func TestTransformFor_ForStmtNotInAST(t *testing.T) {
	parsed, err := parser.Parse("package main\n")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	bogus := parser.ForDirective{
		Node: newCanonicalForStmt(),
	}

	err = transformFor(parsed, bogus)
	if err == nil {
		t.Fatal("expected error when for statement is detached from AST")
	}
	if !strings.Contains(err.Error(), "for statement not found") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestTransform_PropagatesForError verifies that a failing transformFor
// aborts Transform and returns nil, consistent with the other directives.
func TestTransform_PropagatesForError(t *testing.T) {
	parsed, err := parser.Parse("package main\n")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	parsed.Nodes = append(parsed.Nodes, parser.AnnotatedNode{
		Directive: parser.ForDirective{
			Node: &ast.BlockStmt{},
		},
	})

	got, err := Transform(parsed)
	if err == nil {
		t.Fatal("expected Transform to propagate the underlying error")
	}
	if got != nil {
		t.Errorf("expected nil ParseResult on error, got %v", got)
	}
}

// TestFindSchedule_NoneReturnsFalse verifies that findSchedule reports absence
// when no schedule clause is present, so the loop falls back to static.
func TestFindSchedule_NoneReturnsFalse(t *testing.T) {
	if _, ok := findSchedule(nil); ok {
		t.Error("expected findSchedule to return false for empty clause list")
	}
	other := []parser.Clause{parser.PrivateClause{Vars: []string{"x"}}}
	if _, ok := findSchedule(other); ok {
		t.Error("expected findSchedule to return false when no schedule clause present")
	}
}

// TestTransform_ParallelFor_RejectsNonCanonical verifies the canonical-loop
// guard: loops outside OpenMP's canonical form (a step that is not an
// increment/decrement, a condition that does not test the induction variable,
// or a step whose direction contradicts the condition) are rejected with a
// diagnostic instead of being silently miscompiled.
func TestTransform_ParallelFor_RejectsNonCanonical(t *testing.T) {
	cases := []struct {
		name    string
		loop    string
		wantSub string
	}{
		{"non-increment step", "for i := 0; i < N; i *= 2", "step must be"},
		{"wrong cond var", "for i := 0; j < N; i++", "condition must test"},
		{"direction mismatch", "for i := 0; i < N; i--", "direction disagree"},
		{"missing init", "for ; i < N; i++", "init"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			src := "package main\n\nfunc main() {\n\tconst N = 10\n\ti := 0\n\tj := 0\n\t_, _ = i, j\n\t//gompher parallel for\n\t" +
				tc.loop + " {\n\t\t_ = i\n\t}\n}\n"
			parsed, err := parser.Parse(src)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			_, err = Transform(parsed)
			if err == nil {
				t.Fatalf("expected rejection for %q", tc.loop)
			}
			if !strings.Contains(err.Error(), tc.wantSub) {
				t.Errorf("expected error containing %q, got: %v", tc.wantSub, err)
			}
		})
	}
}

// TestTransform_ParallelFor_NormalizesCanonicalForms verifies b1: the accepted
// canonical forms beyond the narrow `i:=0; i<N; i++` are normalized onto the
// runtime's [0, N) space, recovering the induction variable in the body and
// passing the computed trip count.
func TestTransform_ParallelFor_NormalizesCanonicalForms(t *testing.T) {
	cases := []struct {
		name      string
		loop      string
		induction string // expected induction-recovery statement
		count     string // expected trip-count argument
	}{
		{"nonzero start", "for i := 5; i < 10; i++", "i := 5 + _gompherIter", "10-5"},
		{"unit step += 1", "for i := 3; i < 9; i += 1", "i := 3 + _gompherIter", "9-3"},
		{"inclusive <=", "for i := 0; i <= 9; i++", "i := _gompherIter", "9 - 0 + 1"},
		{"stride two", "for i := 0; i < 10; i += 2", "i := _gompherIter * 2", "(10-0+2-1)/2"},
		{"descending --", "for i := 9; i >= 0; i--", "i := 9 - _gompherIter", "9 - 0 + 1"},
		{"descending -= 3", "for i := 9; i > 0; i -= 3", "i := 9 - _gompherIter*3", "(9-0+3-1)/3"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			src := "package main\n\nfunc main() {\n\t//gompher parallel for\n\t" +
				tc.loop + " {\n\t\t_ = i\n\t}\n}\n"
			got := runTransform(t, src)
			// Compare with whitespace removed: go/printer's spacing around an
			// additive chain varies with operand kind (literals compact, idents
			// space out), which is irrelevant to correctness here.
			dense := strings.ReplaceAll(got, " ", "")
			if !strings.Contains(got, "func(_gompherIter int) {") {
				t.Errorf("expected normalized counter param, got:\n%s", got)
			}
			if !strings.Contains(dense, strings.ReplaceAll(tc.induction, " ", "")) {
				t.Errorf("expected induction recovery %q, got:\n%s", tc.induction, got)
			}
			if !strings.Contains(dense, strings.ReplaceAll(tc.count, " ", "")) {
				t.Errorf("expected trip count %q, got:\n%s", tc.count, got)
			}
		})
	}
}

// TestTransform_ParallelFor_SimpleFormUnchanged verifies the fast path: the
// narrow canonical form is emitted with the induction variable as the closure
// parameter and the bare bound as count (no normalization counter).
func TestTransform_ParallelFor_SimpleFormUnchanged(t *testing.T) {
	for _, step := range []string{"i++", "i += 1"} {
		src := "package main\n\nfunc main() {\n\tconst N = 10\n\t//gompher parallel for\n\tfor i := 0; i < N; " +
			step + " {\n\t\t_ = i\n\t}\n}\n"
		got := runTransform(t, src)
		if !strings.Contains(got, "runtime.ParallelFor(func(i int) {") {
			t.Errorf("step %q: expected simple ParallelFor, got:\n%s", step, got)
		}
		if strings.Contains(got, "_gompherIter") {
			t.Errorf("step %q: simple form must not normalize, got:\n%s", step, got)
		}
	}
}

// TestTransform_For_ScheduleStaticChunkedWithClause covers the worksharing
// ForStaticChunked branch of buildWorksharingLoopCall: schedule(static, chunk)
// combined with a data clause routes through the clause expansion.
func TestTransform_For_ScheduleStaticChunkedWithClause(t *testing.T) {
	src := `package main

func main() {
	sum := 0
	//gompher parallel
	{
		//gompher for schedule(static, 4) reduction(+:sum)
		for i := 0; i < 16; i++ {
			sum += i
		}
	}
	_ = sum
}
`
	got := runTransform(t, src)
	if !strings.Contains(got, "runtime.ForStaticChunked(threadID, func(i int) {") {
		t.Errorf("expected worksharing ForStaticChunked with clause, got:\n%s", got)
	}
	if !strings.Contains(got, ", 4)") {
		t.Errorf("expected chunk 4 in the call, got:\n%s", got)
	}
}

// TestTransform_ParallelFor_PrivatePackageQualifiedType covers the package-name
// callback in resolveVarType: a private variable whose type is package-qualified
// (time.Duration) must be rendered with the package's short name.
func TestTransform_ParallelFor_PrivatePackageQualifiedType(t *testing.T) {
	src := `package main

import "time"

func main() {
	d := time.Duration(0)
	//gompher parallel for private(d)
	for i := 0; i < 10; i++ {
		d = time.Duration(i)
		_ = d
	}
	_ = d
}
`
	got := runTransform(t, src)
	if !strings.Contains(got, "var d time.Duration") {
		t.Errorf("expected `var d time.Duration` private decl, got:\n%s", got)
	}
}
