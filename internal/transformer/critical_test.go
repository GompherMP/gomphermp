package transformer

import (
	"go/ast"
	"go/format"
	"strings"
	"testing"

	"github.com/gomphermp/gomphermp/internal/parser"
)

// runTransform is a test helper that parses src, runs the transformer, and
// returns the formatted output.
func runTransform(t *testing.T, src string) string {
	t.Helper()

	parsed, err := parser.Parse(src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	transformed, err := Transform(parsed)
	if err != nil {
		t.Fatalf("transform: %v", err)
	}

	var buf strings.Builder
	if err := format.Node(&buf, transformed.FileSet, transformed.File); err != nil {
		t.Fatalf("format: %v", err)
	}
	return buf.String()
}

// TestTransform_Critical_Anonymous verifies that an anonymous critical
// ("//gompher critical" with no name) becomes a runtime.Critical call whose
// first argument is the empty string.
func TestTransform_Critical_Anonymous(t *testing.T) {
	src := `package main

func main() {
	counter := 0
	//gompher critical
	{
		counter++
	}
	_ = counter
}
`
	got := runTransform(t, src)

	if !strings.Contains(got, `runtime.Critical("", func() {`) {
		t.Errorf("expected runtime.Critical(\"\", func() {...}) in output, got:\n%s", got)
	}
	if !strings.Contains(got, "counter++") {
		t.Errorf("expected original body (counter++) preserved, got:\n%s", got)
	}
}

// TestTransform_Critical_Named verifies that a named critical
// ("//gompher critical(name)") passes its lock name as the first argument to
// runtime.Critical.
func TestTransform_Critical_Named(t *testing.T) {
	src := `package main

func main() {
	x := 0
	//gompher critical(mylock)
	{
		x++
	}
	_ = x
}
`
	got := runTransform(t, src)

	if !strings.Contains(got, `runtime.Critical("mylock", func() {`) {
		t.Errorf("expected runtime.Critical(\"mylock\", func() {...}) in output, got:\n%s", got)
	}
}

// TestTransform_Critical_AddsRuntimeImport verifies that the runtime import
// is injected when a critical directive is present.
func TestTransform_Critical_AddsRuntimeImport(t *testing.T) {
	src := `package main

func main() {
	c := 0
	//gompher critical
	{
		c++
	}
	_ = c
}
`
	got := runTransform(t, src)

	if !strings.Contains(got, `"github.com/gomphermp/gomphermp/pkg/runtime"`) {
		t.Errorf("expected runtime import in output, got:\n%s", got)
	}
}

// TestTransform_Critical_PreservesUserImports verifies that pre-existing
// imports (like fmt) survive the transformation and that the new runtime
// import is added alongside them rather than replacing them.
func TestTransform_Critical_PreservesUserImports(t *testing.T) {
	src := `package main

import "fmt"

func main() {
	c := 0
	//gompher critical
	{
		c++
		fmt.Println(c)
	}
}
`
	got := runTransform(t, src)

	if !strings.Contains(got, `"fmt"`) {
		t.Errorf("expected fmt import preserved, got:\n%s", got)
	}
	if !strings.Contains(got, `"github.com/gomphermp/gomphermp/pkg/runtime"`) {
		t.Errorf("expected runtime import added, got:\n%s", got)
	}
}

// TestTransform_Critical_MultipleInSameFile verifies that two consecutive
// critical directives both get rewritten correctly.
func TestTransform_Critical_MultipleInSameFile(t *testing.T) {
	src := `package main

func main() {
	a, b := 0, 0
	//gompher critical(lockA)
	{
		a++
	}
	//gompher critical(lockB)
	{
		b++
	}
	_, _ = a, b
}
`
	got := runTransform(t, src)

	if !strings.Contains(got, `runtime.Critical("lockA"`) {
		t.Errorf("expected first directive rewritten with lockA, got:\n%s", got)
	}
	if !strings.Contains(got, `runtime.Critical("lockB"`) {
		t.Errorf("expected second directive rewritten with lockB, got:\n%s", got)
	}
}

// TestTransform_Critical_PreservesMultiStmtBody verifies that a body
// containing multiple statements (declaration, computation, side effect) is
// preserved verbatim inside the synthesized closure.
func TestTransform_Critical_PreservesMultiStmtBody(t *testing.T) {
	src := `package main

import "fmt"

func main() {
	total := 0
	//gompher critical
	{
		x := 10
		y := 20
		total = x + y
		fmt.Println(total)
	}
}
`
	got := runTransform(t, src)

	for _, want := range []string{"x := 10", "y := 20", "total = x + y", "fmt.Println(total)"} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q preserved, got:\n%s", want, got)
		}
	}
}

// TestTransform_NoCriticalDirectives_NoRuntimeImport verifies that the
// runtime import is NOT added when no directive that uses the runtime
// appears.
func TestTransform_NoCriticalDirectives_NoRuntimeImport(t *testing.T) {
	src := `package main

import "fmt"

func main() {
	fmt.Println("hello")
}
`
	got := runTransform(t, src)

	if strings.Contains(got, "gomphermp/pkg/runtime") {
		t.Errorf("expected runtime import absent for file with no directives, got:\n%s", got)
	}
}

// TestTransformCritical_WrongNodeType verifies that handing transformCritical
// a directive whose Node is not a *ast.BlockStmt produces an error.
func TestTransformCritical_WrongNodeType(t *testing.T) {
	parsed, err := parser.Parse("package main\n")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	bogus := parser.CriticalDirective{
		Node: &ast.ExprStmt{},
	}

	err = transformCritical(parsed, bogus)
	if err == nil {
		t.Fatal("expected error for non-BlockStmt Node")
	}
	if !strings.Contains(err.Error(), "expected *ast.BlockStmt") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestTransformCritical_BodyNotInAST verifies that when the directive's
// BlockStmt is not actually reachable from the file's AST, the transformer 
// reports the inconsistency.
func TestTransformCritical_BodyNotInAST(t *testing.T) {
	parsed, err := parser.Parse("package main\n")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	detached := &ast.BlockStmt{}
	bogus := parser.CriticalDirective{
		Node: detached,
	}

	err = transformCritical(parsed, bogus)
	if err == nil {
		t.Fatal("expected error when body block is detached from AST")
	}
	if !strings.Contains(err.Error(), "body block not found") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestTransform_PropagatesCriticalError verifies that when transformCritical
// fails, Transform stops and propagates the error to the caller rather than
// returning a partially transformed file.
func TestTransform_PropagatesCriticalError(t *testing.T) {
	parsed, err := parser.Parse("package main\n")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	parsed.Nodes = append(parsed.Nodes, parser.AnnotatedNode{
		Directive: parser.CriticalDirective{
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
