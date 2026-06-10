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
