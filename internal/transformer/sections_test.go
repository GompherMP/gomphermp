package transformer

import (
	"go/ast"
	"strings"
	"testing"

	"github.com/gomphermp/gomphermp/internal/parser"
)

// TestTransform_Sections_BasicRewrite verifies the canonical sections
// transformation: a sections block with three section children becomes a
// single runtime.Sections([]func(){...}) call carrying one closure per
// section, in source order.
func TestTransform_Sections_BasicRewrite(t *testing.T) {
	src := `package main

func main() {
	//gompher sections
	{
		//gompher section
		{
			alpha()
		}
		//gompher section
		{
			bravo()
		}
		//gompher section
		{
			charlie()
		}
	}
}

func alpha() {}
func bravo() {}
func charlie() {}
`
	got := runTransform(t, src)

	if !strings.Contains(got, `runtime.Sections([]func(){`) {
		t.Errorf("expected runtime.Sections([]func(){...}), got:\n%s", got)
	}
	for _, want := range []string{"alpha()", "bravo()", "charlie()"} {
		if !strings.Contains(got, want) {
			t.Errorf("expected section body %q preserved, got:\n%s", want, got)
		}
	}
	// Source order must be preserved: alpha before bravo before charlie. The
	// section names are chosen so none is a substring of "func()" (which would
	// otherwise produce a spurious early match).
	ai, bi, ci := strings.Index(got, "alpha()"), strings.Index(got, "bravo()"), strings.Index(got, "charlie()")
	if !(ai < bi && bi < ci) {
		t.Errorf("expected sections in source order; got positions alpha=%d bravo=%d charlie=%d\n%s", ai, bi, ci, got)
	}
}

// TestTransform_Sections_AddsRuntimeImport verifies the runtime import is
// injected for a sections construct in a file with no prior runtime import.
func TestTransform_Sections_AddsRuntimeImport(t *testing.T) {
	src := `package main

func main() {
	//gompher sections
	{
		//gompher section
		{
			work()
		}
	}
}

func work() {}
`
	got := runTransform(t, src)

	if !strings.Contains(got, `"github.com/gomphermp/gomphermp/pkg/runtime"`) {
		t.Errorf("expected runtime import in output, got:\n%s", got)
	}
}

// TestTransform_Sections_RemovesDirectiveComments verifies that neither the
// //gompher sections comment nor any //gompher section comment survives in the
// output. Orphaned directive comments would render in awkward positions and
// signal that the construct was only partially consumed.
func TestTransform_Sections_RemovesDirectiveComments(t *testing.T) {
	src := `package main

func main() {
	//gompher sections
	{
		//gompher section
		{
			a()
		}
		//gompher section
		{
			b()
		}
	}
}

func a() {}
func b() {}
`
	got := runTransform(t, src)

	if strings.Contains(got, "//gompher") {
		t.Errorf("expected all //gompher comments removed, got:\n%s", got)
	}
}

// TestTransform_Sections_PreservesMultiStmtBody verifies that a section whose
// body has multiple statements is wrapped verbatim, with no analysis or
// reordering of the inner statements.
func TestTransform_Sections_PreservesMultiStmtBody(t *testing.T) {
	src := `package main

import "fmt"

func main() {
	//gompher sections
	{
		//gompher section
		{
			x := 1
			y := 2
			fmt.Println(x + y)
		}
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

// TestTransformSections_WrongNodeType verifies the defensive error path when
// the directive's Node is not a *ast.BlockStmt.
func TestTransformSections_WrongNodeType(t *testing.T) {
	parsed, err := parser.Parse("package main\n")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	bogus := parser.SectionsDirective{
		Node: &ast.ExprStmt{},
	}

	err = transformSections(parsed, bogus)
	if err == nil {
		t.Fatal("expected error for non-BlockStmt Node")
	}
	if !strings.Contains(err.Error(), "expected *ast.BlockStmt") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestTransformSections_NoSectionBlocks verifies that a sections block whose
// inner blocks are not registered sections (so none match) is rejected with a
// descriptive error. This also exercises findSectionDirective's not-found
// branch: the inner block has no matching SectionDirective and is skipped.
func TestTransformSections_NoSectionBlocks(t *testing.T) {
	parsed, err := parser.Parse("package main\n\nfunc main() {}\n")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	inner := &ast.BlockStmt{}
	outer := &ast.BlockStmt{List: []ast.Stmt{inner}}
	d := parser.SectionsDirective{Node: outer}

	err = transformSections(parsed, d)
	if err == nil {
		t.Fatal("expected error when no section blocks are present")
	}
	if !strings.Contains(err.Error(), "no section blocks found") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestTransformSections_BodyNotInAST verifies that when the sections block has
// genuine section children but the outer block is not reachable from the file
// AST, the transformer reports the inconsistency. This exercises the
// replaceBlockStmt-failure branch with findSectionDirective succeeding.
func TestTransformSections_BodyNotInAST(t *testing.T) {
	parsed, err := parser.Parse("package main\n\nfunc main() {}\n")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	inner := &ast.BlockStmt{}
	// Include a non-block statement before the section so the loop exercises
	// its "skip non-BlockStmt" branch in addition to the failure path.
	stray := &ast.ExprStmt{X: &ast.Ident{Name: "x"}}
	outer := &ast.BlockStmt{List: []ast.Stmt{stray, inner}}

	// Register inner as a real section so findSectionDirective matches it.
	parsed.Nodes = append(parsed.Nodes, parser.AnnotatedNode{
		Directive: parser.SectionDirective{Node: inner},
	})

	d := parser.SectionsDirective{Node: outer}

	err = transformSections(parsed, d)
	if err == nil {
		t.Fatal("expected error when outer block is detached from AST")
	}
	if !strings.Contains(err.Error(), "body block not found") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestTransform_PropagatesSectionsError verifies that a failing
// transformSections aborts Transform and returns nil, consistent with the
// other directives.
func TestTransform_PropagatesSectionsError(t *testing.T) {
	parsed, err := parser.Parse("package main\n")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	parsed.Nodes = append(parsed.Nodes, parser.AnnotatedNode{
		Directive: parser.SectionsDirective{
			Node: &ast.ExprStmt{},
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
