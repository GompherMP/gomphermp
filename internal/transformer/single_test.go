package transformer

import (
	"go/ast"
	"strings"
	"testing"

	"github.com/gomphermp/gomphermp/internal/parser"
)

// TestTransform_Single_BasicRewrite verifies the canonical single
// transformation: //gompher single { body } becomes runtime.Single(func() {
// body }). This is the baseline contract for the directive.
func TestTransform_Single_BasicRewrite(t *testing.T) {
	src := `package main

func main() {
	//gompher single
	{
		setup()
	}
	_ = 0
}

func setup() {}
`
	got := runTransform(t, src)

	if !strings.Contains(got, `runtime.Single(func() {`) {
		t.Errorf("expected runtime.Single(func() {...}) in output, got:\n%s", got)
	}
	if !strings.Contains(got, "setup()") {
		t.Errorf("expected original body call preserved, got:\n%s", got)
	}
}

// TestTransform_Single_AddsRuntimeImport verifies that the runtime import is
// injected when a single directive appears in a file with no prior imports.
// Without this, the synthesized runtime.Single call would reference an
// undefined identifier and the file would not compile.
func TestTransform_Single_AddsRuntimeImport(t *testing.T) {
	src := `package main

func main() {
	//gompher single
	{
		work()
	}
}

func work() {}
`
	got := runTransform(t, src)

	if !strings.Contains(got, `"github.com/gomphermp/gomphermp/pkg/runtime"`) {
		t.Errorf("expected runtime import in output, got:\n%s", got)
	}
}

// TestTransform_Single_PreservesMultiStmtBody verifies that a body with
// several statements is moved verbatim into the closure. Single's body in
// OpenMP can be arbitrary code; the transformer must not analyze or rewrite
// it in this phase.
func TestTransform_Single_PreservesMultiStmtBody(t *testing.T) {
	src := `package main

import "fmt"

func main() {
	//gompher single
	{
		a := 1
		b := 2
		fmt.Println(a + b)
	}
}
`
	got := runTransform(t, src)

	for _, want := range []string{"a := 1", "b := 2", "fmt.Println(a + b)"} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q preserved, got:\n%s", want, got)
		}
	}
}

// TestTransformSingle_WrongNodeType verifies the defensive error path: when
// the directive's Node is not a *ast.BlockStmt (which should not happen with
// a healthy parser), transformSingle returns a descriptive error instead of
// panicking. Critical, Single and Master share this guard via
// transformBlockDirective; this test exercises it through the single entry
// point so the contract is checked end-to-end.
func TestTransformSingle_WrongNodeType(t *testing.T) {
	parsed, err := parser.Parse("package main\n")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	bogus := parser.SingleDirective{
		Node: &ast.ExprStmt{},
	}

	err = transformSingle(parsed, bogus)
	if err == nil {
		t.Fatal("expected error for non-BlockStmt Node")
	}
	if !strings.Contains(err.Error(), "expected *ast.BlockStmt") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestTransform_PropagatesSingleError verifies that when transformSingle
// fails, Transform propagates the error and returns nil instead of a
// partially transformed file. Mirror of the equivalent contract test for
// Critical, ensuring the error-handling branch in Transform's dispatch is
// exercised for SingleDirective.
func TestTransform_PropagatesSingleError(t *testing.T) {
	parsed, err := parser.Parse("package main\n")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	parsed.Nodes = append(parsed.Nodes, parser.AnnotatedNode{
		Directive: parser.SingleDirective{
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
