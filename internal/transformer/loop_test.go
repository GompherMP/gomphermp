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

	if !strings.Contains(got, `runtime.For(func(i int) {`) {
		t.Errorf("expected runtime.For(func(i int) {...}, N), got:\n%s", got)
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
// c) clause redirects the combined construct to runtime.ForDynamic with the
// requested chunk size, rather than the static ParallelFor entry point.
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

	if !strings.Contains(got, `runtime.ForDynamic(func(i int) {`) {
		t.Errorf("expected runtime.ForDynamic for schedule(dynamic), got:\n%s", got)
	}
	if !strings.Contains(got, "}, N, 4)") {
		t.Errorf("expected bound N and chunk 4 as trailing args, got:\n%s", got)
	}
	if strings.Contains(got, "ParallelFor") {
		t.Errorf("expected ParallelFor NOT emitted when schedule is dynamic, got:\n%s", got)
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
// schedule(static) does NOT divert to ForDynamic — static scheduling keeps
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

	if !strings.Contains(got, `runtime.For(func(i int) {`) {
		t.Errorf("expected runtime.For for schedule(static), got:\n%s", got)
	}
	if strings.Contains(got, "ForDynamic") {
		t.Errorf("expected ForDynamic NOT emitted for static schedule, got:\n%s", got)
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

	if !strings.Contains(got, `runtime.For(func(idx int) {`) {
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

// TestTransformFor_NonCanonicalCondition verifies that a loop whose condition
// is not the canonical i < N form (here it uses <=) is rejected with a
// descriptive error. extractUpperBound only accepts the '<' operator, so this
// exercises the bound-extraction error branch of transformLoopDirective.
func TestTransformFor_NonCanonicalCondition(t *testing.T) {
	src := `package main

func main() {
	const N = 10
	//gompher for
	for i := 0; i <= N; i++ {
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
		t.Fatal("expected error for non-canonical loop condition (<=)")
	}
	if !strings.Contains(err.Error(), "'<'") {
		t.Errorf("expected error mentioning the '<' operator requirement, got: %v", err)
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
