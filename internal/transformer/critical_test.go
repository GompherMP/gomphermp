package transformer

import (
	"go/format"
	"strings"
	"testing"

	"github.com/gomphermp/gomphermp/internal/parser"
)

// runTransform is a test helper that parses src, runs the transformer, and
// returns the formatted output. Failures in parse/transform/format abort the
// test directly so individual cases can focus on asserting against the
// output string.
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
// (`//gompher critical` with no name) becomes a runtime.Critical call whose
// first argument is the empty string. This is the most common form and the
// baseline contract for the directive.
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
// (`//gompher critical(name)`) passes its lock name as the first argument to
// runtime.Critical. The runtime treats distinct names as independent locks,
// so getting the name through correctly is what enables fine-grained mutual
// exclusion in transpiled programs.
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
// is injected when a critical directive is present. Without this, the
// transformed file would reference an undefined `runtime` identifier and
// fail to compile.
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
// critical directives both get rewritten correctly. Earlier phases of the
// transformer iterated through parser.Nodes in order; this test confirms the
// iteration does not skip or re-process entries when each one mutates the
// AST in place.
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
// preserved verbatim inside the synthesized closure. Bodies are not analyzed
// in this phase — they are moved wholesale into the closure — so the test
// confirms there is no accidental reformatting or statement dropping.
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
// appears. This is the complement of AddsRuntimeImport and guards against
// accidentally polluting files that have nothing to transform.
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
