package transformer

import (
	"go/ast"
	"go/token"
	"strings"
	"testing"

	"github.com/gomphermp/gomphermp/internal/parser"
)

// TestTransform_Taskloop_BasicRewrite verifies the canonical taskloop
// transformation: //gompher taskloop for i := 0; i < N; i++ { body }
// becomes runtime.Taskloop(func(i int) { body }, N, 1).
func TestTransform_Taskloop_BasicRewrite(t *testing.T) {
	src := `package main

func main() {
	n := 10
	//gompher taskloop
	for i := 0; i < n; i++ {
		work(i)
	}
}

func work(i int) {}
`
	got := runTransform(t, src)

	if !strings.Contains(got, "runtime.Taskloop(func(i int) {") {
		t.Errorf("expected runtime.Taskloop(func(i int) {...}) in output, got:\n%s", got)
	}
	if !strings.Contains(got, "work(i)") {
		t.Errorf("expected body call preserved, got:\n%s", got)
	}
}

// TestTransform_Taskloop_WithGrainsize verifies that grainsize(4) produces
// 4 as the third argument to Taskloop.
func TestTransform_Taskloop_WithGrainsize(t *testing.T) {
	src := `package main

func main() {
	n := 100
	//gompher taskloop grainsize(4)
	for i := 0; i < n; i++ {
		work(i)
	}
}

func work(i int) {}
`
	got := runTransform(t, src)

	if !strings.Contains(got, "runtime.Taskloop(") {
		t.Errorf("expected runtime.Taskloop in output, got:\n%s", got)
	}
	if !strings.Contains(got, ", 4)") {
		t.Errorf("expected grainsize 4 as last argument, got:\n%s", got)
	}
}

// TestTransform_Taskloop_DefaultGrainsize verifies that a taskloop without a
// grainsize clause uses 1 as the default.
func TestTransform_Taskloop_DefaultGrainsize(t *testing.T) {
	src := `package main

func main() {
	n := 50
	//gompher taskloop
	for i := 0; i < n; i++ {
		process(i)
	}
}

func process(i int) {}
`
	got := runTransform(t, src)

	if !strings.Contains(got, ", 1)") {
		t.Errorf("expected default grainsize 1 as last argument, got:\n%s", got)
	}
}

// TestTransform_Taskloop_AddsRuntimeImport verifies that the runtime import
// is injected when a taskloop directive appears in a file with no prior imports.
func TestTransform_Taskloop_AddsRuntimeImport(t *testing.T) {
	src := `package main

func main() {
	n := 10
	//gompher taskloop
	for i := 0; i < n; i++ {
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

// TestTransform_Taskloop_PreservesMultiStmtBody verifies that a loop body
// with several statements is preserved verbatim inside the closure.
func TestTransform_Taskloop_PreservesMultiStmtBody(t *testing.T) {
	src := `package main

import "fmt"

func main() {
	n := 20
	//gompher taskloop
	for i := 0; i < n; i++ {
		x := i * 2
		y := x + 1
		fmt.Println(y)
	}
}
`
	got := runTransform(t, src)

	for _, want := range []string{"x := i * 2", "y := x + 1", "fmt.Println(y)"} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q preserved, got:\n%s", want, got)
		}
	}
}

// TestTransform_Taskloop_LiteralBound verifies that a literal integer bound
// is preserved as-is in the emitted call.
func TestTransform_Taskloop_LiteralBound(t *testing.T) {
	src := `package main

func main() {
	//gompher taskloop
	for i := 0; i < 64; i++ {
		work(i)
	}
}

func work(i int) {}
`
	got := runTransform(t, src)

	if !strings.Contains(got, "64") {
		t.Errorf("expected literal bound 64 in output, got:\n%s", got)
	}
}

// TestTransformTaskloop_WrongNodeType verifies the defensive error path when
// the directive's Node is not a *ast.ForStmt.
func TestTransformTaskloop_WrongNodeType(t *testing.T) {
	parsed, err := parser.Parse("package main\n")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	bogus := parser.TaskloopDirective{
		Node: &ast.BlockStmt{},
	}

	err = transformTaskloop(parsed, bogus)
	if err == nil {
		t.Fatal("expected error for non-ForStmt Node")
	}
	if !strings.Contains(err.Error(), "expected *ast.ForStmt") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestTransformTaskloop_NoInitStatement verifies that a for loop without an
// init statement produces a descriptive error.
func TestTransformTaskloop_NoInitStatement(t *testing.T) {
	parsed, err := parser.Parse("package main\n")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	bogus := parser.TaskloopDirective{
		Node: &ast.ForStmt{
			Body: &ast.BlockStmt{},
		},
	}

	err = transformTaskloop(parsed, bogus)
	if err == nil {
		t.Fatal("expected error for missing init statement")
	}
	if !strings.Contains(err.Error(), "no init statement") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestTransformTaskloop_NonBinaryCondition verifies that a for loop whose
// condition is not a BinaryExpr (e.g. a bare identifier) returns a descriptive
// error rather than panicking.
func TestTransformTaskloop_NonBinaryCondition(t *testing.T) {
	parsed, err := parser.Parse("package main\n")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	bogus := parser.TaskloopDirective{
		Node: &ast.ForStmt{
			Init: &ast.AssignStmt{
				Lhs: []ast.Expr{&ast.Ident{Name: "i"}},
				Rhs: []ast.Expr{&ast.BasicLit{}},
			},
			Cond: &ast.Ident{Name: "cond"}, // not a BinaryExpr
			Body: &ast.BlockStmt{},
		},
	}

	err = transformTaskloop(parsed, bogus)
	if err == nil {
		t.Fatal("expected error for non-BinaryExpr condition")
	}
	if !strings.Contains(err.Error(), "binary expression") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestTransformTaskloop_WrongOperator verifies that a BinaryExpr condition
// using an operator other than '<' (e.g. '<=') returns a descriptive error.
// The code path checks binExpr.Op != token.LSS specifically.
func TestTransformTaskloop_WrongOperator(t *testing.T) {
	parsed, err := parser.Parse("package main\n")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	bogus := parser.TaskloopDirective{
		Node: &ast.ForStmt{
			Init: &ast.AssignStmt{
				Tok: token.DEFINE,
				Lhs: []ast.Expr{&ast.Ident{Name: "i"}},
				Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "0"}},
			},
			Cond: &ast.BinaryExpr{
				Op: token.LEQ, // <= instead of <
				X:  &ast.Ident{Name: "i"},
				Y:  &ast.Ident{Name: "n"},
			},
			Body: &ast.BlockStmt{},
		},
	}

	err = transformTaskloop(parsed, bogus)
	if err == nil {
		t.Fatal("expected error for non-LSS operator in for condition")
	}
	if !strings.Contains(err.Error(), "'<'") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestTransformTaskloop_ForStmtNotInAST verifies that when the directive's
// ForStmt has a valid shape (passes extractLoopVar and extractUpperBound) but
// is not reachable from the file's AST, the transformer reports the
// inconsistency rather than silently succeeding.
func TestTransformTaskloop_ForStmtNotInAST(t *testing.T) {
	parsed, err := parser.Parse("package main\n")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	detached := &ast.ForStmt{
		Init: &ast.AssignStmt{
			Tok: token.DEFINE,
			Lhs: []ast.Expr{&ast.Ident{Name: "i"}},
			Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "0"}},
		},
		Cond: &ast.BinaryExpr{
			Op: token.LSS,
			X:  &ast.Ident{Name: "i"},
			Y:  &ast.Ident{Name: "n"},
		},
		Body: &ast.BlockStmt{},
	}

	bogus := parser.TaskloopDirective{Node: detached}

	err = transformTaskloop(parsed, bogus)
	if err == nil {
		t.Fatal("expected error when ForStmt is detached from AST")
	}
	if !strings.Contains(err.Error(), "for statement not found") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestTransform_Taskloop_DifferentLoopVariable verifies that the loop variable
// name is faithfully propagated into the closure signature and body when it
// is not the conventional 'i'.
func TestTransform_Taskloop_DifferentLoopVariable(t *testing.T) {
	src := `package main

func main() {
	n := 10
	//gompher taskloop
	for j := 0; j < n; j++ {
		work(j)
	}
}

func work(j int) {}
`
	got := runTransform(t, src)

	if !strings.Contains(got, "func(j int) {") {
		t.Errorf("expected loop variable j in closure signature, got:\n%s", got)
	}
	if !strings.Contains(got, "work(j)") {
		t.Errorf("expected body with j preserved, got:\n%s", got)
	}
}

// TestTransform_Taskloop_DirectiveCommentRemoved verifies that the
// //gompher taskloop comment is stripped from the output after transformation.
func TestTransform_Taskloop_DirectiveCommentRemoved(t *testing.T) {
	src := `package main

func main() {
	n := 10
	//gompher taskloop
	for i := 0; i < n; i++ {
		work(i)
	}
}

func work(i int) {}
`
	got := runTransform(t, src)

	if strings.Contains(got, "//gompher") {
		t.Errorf("expected directive comment removed from output, got:\n%s", got)
	}
}

// TestTransform_PropagatesTaskloopError verifies that Transform propagates
// taskloop errors and returns nil for a partially transformed file.
func TestTransform_PropagatesTaskloopError(t *testing.T) {
	parsed, err := parser.Parse("package main\n")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	parsed.Nodes = append(parsed.Nodes, parser.AnnotatedNode{
		Directive: parser.TaskloopDirective{
			Node: &ast.ForStmt{Body: &ast.BlockStmt{}},
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
