package transformer

import (
	"go/ast"
	"strings"
	"testing"

	"github.com/gomphermp/gomphermp/internal/parser"
)

// TestTransform_Parallel_BasicRewrite verifies the canonical parallel
// transformation: //gompher parallel { body } becomes
// runtime.Parallel(func(threadID int) { body }). This is the baseline
// contract for the directive and the foundation for nested master/single.
func TestTransform_Parallel_BasicRewrite(t *testing.T) {
	src := `package main

import "fmt"

func main() {
	//gompher parallel
	{
		fmt.Println("hello from the team")
	}
	fmt.Println("done")
}
`
	got := runTransform(t, src)

	if !strings.Contains(got, `runtime.Parallel(func(threadID int) {`) {
		t.Errorf("expected runtime.Parallel(func(threadID int) {...}) in output, got:\n%s", got)
	}
	if !strings.Contains(got, `fmt.Println("hello from the team")`) {
		t.Errorf("expected original body preserved, got:\n%s", got)
	}
}

// TestTransform_Parallel_AddsRuntimeImport verifies that the runtime import
// is injected when a parallel directive appears in a file whose only import
// is the user's fmt. Without it the synthesized runtime.Parallel call would
// not compile.
func TestTransform_Parallel_AddsRuntimeImport(t *testing.T) {
	src := `package main

import "fmt"

func main() {
	//gompher parallel
	{
		fmt.Println("work")
	}
}
`
	got := runTransform(t, src)

	if !strings.Contains(got, `"github.com/gomphermp/gomphermp/pkg/runtime"`) {
		t.Errorf("expected runtime import in output, got:\n%s", got)
	}
}

// TestTransform_Parallel_BodyUsesThreadID verifies that a body referencing
// threadID compiles cleanly after transformation: the identifier emitted by
// the transformer must match the one the body expects. This is the contract
// that lets users branch on the goroutine index inside a parallel region.
func TestTransform_Parallel_BodyUsesThreadID(t *testing.T) {
	src := `package main

import "fmt"

func main() {
	//gompher parallel
	{
		fmt.Println("thread", threadID)
	}
}
`
	got := runTransform(t, src)

	if !strings.Contains(got, `func(threadID int) {`) {
		t.Errorf("expected threadID parameter in closure, got:\n%s", got)
	}
	if !strings.Contains(got, "fmt.Println(\"thread\", threadID)") {
		t.Errorf("expected body reference to threadID preserved, got:\n%s", got)
	}
}

// TestTransform_Parallel_EnclosesMaster verifies that, after
// transformation, runtime.Master(threadID, ...) must sit inside the
// runtime.Parallel(func(threadID int) {...}) closure, so the threadID
// reference resolves to the closure parameter rather than an undefined name.
func TestTransform_Parallel_EnclosesMaster(t *testing.T) {
	src := `package main

import "fmt"

func main() {
	//gompher parallel
	{
		//gompher master
		{
			fmt.Println("only thread 0")
		}
	}
}
`
	got := runTransform(t, src)

	if !strings.Contains(got, `runtime.Parallel(func(threadID int) {`) {
		t.Errorf("expected outer Parallel closure, got:\n%s", got)
	}
	if !strings.Contains(got, `runtime.Master(threadID, func() {`) {
		t.Errorf("expected nested Master call referencing threadID, got:\n%s", got)
	}

	// The Master call must appear after the Parallel call in the output, i.e.
	// nested inside it, not as a sibling.
	pIdx := strings.Index(got, "runtime.Parallel")
	mIdx := strings.Index(got, "runtime.Master")
	if pIdx == -1 || mIdx == -1 || mIdx < pIdx {
		t.Errorf("expected Master nested inside Parallel, got:\n%s", got)
	}
}

// TestTransform_Parallel_PreservesMultiStmtBody verifies that a body with
// several statements is moved verbatim into the closure, with no analysis or
// reordering.
func TestTransform_Parallel_PreservesMultiStmtBody(t *testing.T) {
	src := `package main

import "fmt"

func main() {
	//gompher parallel
	{
		x := 1
		y := 2
		fmt.Println(x + y)
	}
}
`
	got := runTransform(t, src)

	for _, want := range []string{"x := 1", "y := 2", "fmt.Println(x + y)"} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q preserved, got:\n%s", want, got)
		}
	}
}

// TestTransform_Parallel_Private verifies that a private(x) clause prepends a
// fresh `var x T` declaration inside the closure, giving each goroutine its own
// zero-valued copy that shadows the outer variable.
func TestTransform_Parallel_Private(t *testing.T) {
	src := `package main

func main() {
	var local int
	//gompher parallel private(local)
	{
		local = 1
		_ = local
	}
	_ = local
}
`
	got := runTransform(t, src)

	if !strings.Contains(got, "runtime.Parallel(func(threadID int) {") {
		t.Fatalf("expected Parallel closure, got:\n%s", got)
	}
	if !strings.Contains(got, "var local int") {
		t.Errorf("expected private declaration `var local int` inside closure, got:\n%s", got)
	}
	// The private decl must be inside the closure, after the Parallel call opens.
	pIdx := strings.Index(got, "runtime.Parallel")
	dIdx := strings.LastIndex(got, "var local int")
	if dIdx < pIdx {
		t.Errorf("expected private decl inside the closure, got:\n%s", got)
	}
}

// TestTransform_Parallel_Firstprivate verifies that firstprivate(x) snapshots
// the outer value before the call (_fp_x := x) and shadows it inside the
// closure (x := _fp_x), so each goroutine starts from the captured value.
func TestTransform_Parallel_Firstprivate(t *testing.T) {
	src := `package main

func main() {
	var base int = 10
	//gompher parallel firstprivate(base)
	{
		_ = base
	}
	_ = base
}
`
	got := runTransform(t, src)

	if !strings.Contains(got, "_fp_base := base") {
		t.Errorf("expected firstprivate capture `_fp_base := base` before the call, got:\n%s", got)
	}
	if !strings.Contains(got, "base := _fp_base") {
		t.Errorf("expected firstprivate shadow `base := _fp_base` inside the closure, got:\n%s", got)
	}
	// Capture must come before the Parallel call; shadow must come after.
	capIdx := strings.Index(got, "_fp_base := base")
	parIdx := strings.Index(got, "runtime.Parallel")
	shIdx := strings.Index(got, "base := _fp_base")
	if !(capIdx < parIdx && parIdx < shIdx) {
		t.Errorf("expected capture before call and shadow inside; got cap=%d par=%d sh=%d\n%s", capIdx, parIdx, shIdx, got)
	}
}

// TestTransform_Parallel_Shared verifies that shared(x) is a no-op: no
// private declaration or firstprivate capture is emitted, since Go closures
// already share captured variables by reference.
func TestTransform_Parallel_Shared(t *testing.T) {
	src := `package main

func main() {
	var total int
	//gompher parallel shared(total)
	{
		total = 1
		_ = total
	}
	_ = total
}
`
	got := runTransform(t, src)

	if strings.Contains(got, "var total int\n\t\tvar") || strings.Contains(got, "_fp_total") {
		t.Errorf("expected shared(total) to emit no extra declarations, got:\n%s", got)
	}
	if !strings.Contains(got, "runtime.Parallel(func(threadID int) {") {
		t.Errorf("expected a plain Parallel closure, got:\n%s", got)
	}
}

// TestTransform_Parallel_PrivateShortDecl verifies that private(x) on a
// variable declared with := resolves its type via go/types and emits the right
// `var x T` declaration, without requiring an explicit type annotation.
func TestTransform_Parallel_PrivateShortDecl(t *testing.T) {
	src := `package main

func main() {
	x := 5
	//gompher parallel private(x)
	{
		x = 1
		_ = x
	}
	_ = x
}
`
	got := runTransform(t, src)
	if !strings.Contains(got, "var x int") {
		t.Errorf("expected private(x) on `x := 5` to resolve to `var x int`, got:\n%s", got)
	}
}

// TestTransform_Parallel_PrivateUndeclared verifies that private(x) on a
// variable that is not declared before the directive produces a descriptive
// error rather than emitting an untyped declaration.
func TestTransform_Parallel_PrivateUndeclared(t *testing.T) {
	parsed, err := parser.Parse("package main\n\nfunc main() {}\n")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	// Construct a parallel directive that claims private(ghost) where ghost is
	// never declared, so go/types cannot type it.
	parsed.Nodes = append(parsed.Nodes, parser.AnnotatedNode{
		Directive: parser.ParallelDirective{
			Clauses: []parser.Clause{parser.PrivateClause{Vars: []string{"ghost"}}},
			Node:    &ast.BlockStmt{},
		},
	})

	_, err = Transform(parsed)
	if err == nil || !strings.Contains(err.Error(), "private(ghost)") {
		t.Errorf("expected private type-resolution error for undeclared var, got: %v", err)
	}
}

// TestTransformParallel_WrongNodeType verifies the defensive error path:
// passing a non-BlockStmt Node yields a descriptive error rather than a
// panic, consistent with the other directive handlers.
func TestTransformParallel_WrongNodeType(t *testing.T) {
	parsed, err := parser.Parse("package main\n")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	bogus := parser.ParallelDirective{
		Node: &ast.ExprStmt{},
	}

	err = transformParallel(parsed, bogus)
	if err == nil {
		t.Fatal("expected error for non-BlockStmt Node")
	}
	if !strings.Contains(err.Error(), "expected *ast.BlockStmt") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestTransformParallel_BodyNotInAST verifies that a directive whose body is
// not reachable from the file AST produces an error instead of silently
// dropping the runtime call. Mirrors the equivalent guard in critical.
func TestTransformParallel_BodyNotInAST(t *testing.T) {
	parsed, err := parser.Parse("package main\n")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	bogus := parser.ParallelDirective{
		Node: &ast.BlockStmt{},
	}

	err = transformParallel(parsed, bogus)
	if err == nil {
		t.Fatal("expected error when body block is detached from AST")
	}
	if !strings.Contains(err.Error(), "body block not found") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestTransform_PropagatesParallelError verifies that a failing
// transformParallel aborts Transform and returns nil, consistent with the
// error-propagation contract checked for the other directives.
func TestTransform_PropagatesParallelError(t *testing.T) {
	parsed, err := parser.Parse("package main\n")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	parsed.Nodes = append(parsed.Nodes, parser.AnnotatedNode{
		Directive: parser.ParallelDirective{
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
