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
// several statements is moved verbatim into the closure.
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
// the directive's Node is not a *ast.BlockStmt, transformSingle returns 
// a descriptive error instead of panicking.
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
// partially transformed file.
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

// TestTransform_Single_Private verifies that private(x) on a single declares a
// fresh per-region copy inside the Single closure, shadowing the outer one.
func TestTransform_Single_Private(t *testing.T) {
	src := `package main

func main() {
	x := 0
	//gompher parallel
	{
		//gompher single private(x)
		{
			x = 1
			_ = x
		}
	}
	_ = x
}
`
	got := runTransform(t, src)
	if !strings.Contains(got, "runtime.Single(func() {") {
		t.Fatalf("expected Single call, got:\n%s", got)
	}
	// The private declaration must sit inside the Single closure, before the body.
	singleIdx := strings.Index(got, "runtime.Single")
	declIdx := strings.Index(got, "var x int")
	bodyIdx := strings.Index(got, "x = 1")
	if !(singleIdx < declIdx && declIdx < bodyIdx) {
		t.Errorf("expected `var x int` between Single and body; single=%d decl=%d body=%d\n%s", singleIdx, declIdx, bodyIdx, got)
	}
}

// TestTransform_Single_Firstprivate verifies that firstprivate(y) captures the
// outer value before the call and shadows y with it inside the closure.
func TestTransform_Single_Firstprivate(t *testing.T) {
	src := `package main

func main() {
	y := 7
	//gompher parallel
	{
		//gompher single firstprivate(y)
		{
			y = y + 1
			_ = y
		}
	}
	_ = y
}
`
	got := runTransform(t, src)
	for _, want := range []string{
		"_fp_y := y",                   // capture before the call
		"runtime.Single(func() {",      // the single closure
		"y := _fp_y",                   // shadow with the captured value
	} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in firstprivate single output, got:\n%s", want, got)
		}
	}
	// The capture must precede the Single call.
	if capIdx, singleIdx := strings.Index(got, "_fp_y := y"), strings.Index(got, "runtime.Single"); !(capIdx >= 0 && capIdx < singleIdx) {
		t.Errorf("expected capture before Single; cap=%d single=%d\n%s", capIdx, singleIdx, got)
	}
}

// TestTransform_Single_PrivateUndeclared verifies that a private clause naming a
// variable whose type cannot be resolved surfaces a descriptive error.
func TestTransform_Single_PrivateUndeclared(t *testing.T) {
	parsed, err := parser.Parse("package main\n\nfunc main() {}\n")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	parsed.Nodes = append(parsed.Nodes, parser.AnnotatedNode{
		Directive: parser.SingleDirective{
			Clauses: []parser.Clause{parser.PrivateClause{Vars: []string{"ghost"}}},
			Node:    &ast.BlockStmt{},
		},
	})
	if _, err := Transform(parsed); err == nil || !strings.Contains(err.Error(), "private(ghost)") {
		t.Errorf("expected private resolution error, got: %v", err)
	}
}
