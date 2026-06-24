package transformer

import (
	"go/ast"
	"go/token"
	"strings"
	"testing"

	"github.com/gomphermp/gomphermp/internal/parser"
)

// transformErr parses src, runs Transform, and returns the resulting error
// (which may be nil). Unlike runTransform it does not fatal on a transform
// error, so it can drive the directive's error paths through real source.
func transformErr(t *testing.T, src string) error {
	t.Helper()
	parsed, err := parser.Parse(src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	_, err = Transform(parsed)
	return err
}

// ---------------------------------------------------------------------------
// Happy paths (real source, end to end)
// ---------------------------------------------------------------------------

// TestTransform_Atomic_UpdateIncrement verifies x++ becomes an atomic add of 1.
func TestTransform_Atomic_UpdateIncrement(t *testing.T) {
	src := `package main

func main() {
	var counter int
	//gompher atomic update
	counter++
	_ = counter
}
`
	got := runTransform(t, src)
	if !strings.Contains(got, "runtime.AtomicAddInt(&counter, 1)") {
		t.Errorf("expected AtomicAddInt(&counter, 1), got:\n%s", got)
	}
}

// TestTransform_Atomic_UpdateDecrement verifies x-- becomes an atomic add of -1.
func TestTransform_Atomic_UpdateDecrement(t *testing.T) {
	src := `package main

func main() {
	var counter int
	//gompher atomic update
	counter--
	_ = counter
}
`
	got := runTransform(t, src)
	if !strings.Contains(got, "runtime.AtomicAddInt(&counter, -1)") {
		t.Errorf("expected AtomicAddInt(&counter, -1), got:\n%s", got)
	}
}

// TestTransform_Atomic_UpdatePlusAssign verifies x += e becomes an atomic add
// of e.
func TestTransform_Atomic_UpdatePlusAssign(t *testing.T) {
	src := `package main

func main() {
	var total int
	//gompher atomic update
	total += 5
	_ = total
}
`
	got := runTransform(t, src)
	if !strings.Contains(got, "runtime.AtomicAddInt(&total, 5)") {
		t.Errorf("expected AtomicAddInt(&total, 5), got:\n%s", got)
	}
}

// TestTransform_Atomic_UpdateMinusAssign verifies x -= e becomes an atomic add
// of the negated, parenthesized right-hand side.
func TestTransform_Atomic_UpdateMinusAssign(t *testing.T) {
	src := `package main

func main() {
	var total int
	//gompher atomic update
	total -= 3 + 1
	_ = total
}
`
	got := runTransform(t, src)
	if !strings.Contains(got, "runtime.AtomicAddInt(&total, -(3 + 1))") {
		t.Errorf("expected AtomicAddInt(&total, -(3 + 1)), got:\n%s", got)
	}
}

// TestTransform_Atomic_DefaultModeIsUpdate verifies that an atomic directive
// with no explicit mode is treated as update.
func TestTransform_Atomic_DefaultModeIsUpdate(t *testing.T) {
	src := `package main

func main() {
	var counter int
	//gompher atomic
	counter++
	_ = counter
}
`
	got := runTransform(t, src)
	if !strings.Contains(got, "runtime.AtomicAddInt(&counter, 1)") {
		t.Errorf("expected default mode to behave as update, got:\n%s", got)
	}
}

// TestTransform_Atomic_Write verifies x = e becomes an atomic store.
func TestTransform_Atomic_Write(t *testing.T) {
	src := `package main

func main() {
	var flag int
	//gompher atomic write
	flag = 1
	_ = flag
}
`
	got := runTransform(t, src)
	if !strings.Contains(got, "runtime.AtomicStoreInt(&flag, 1)") {
		t.Errorf("expected AtomicStoreInt(&flag, 1), got:\n%s", got)
	}
}

// TestTransform_Atomic_Read verifies v = x replaces only the right-hand side
// with an atomic load, leaving the destination assignment intact.
func TestTransform_Atomic_Read(t *testing.T) {
	src := `package main

func main() {
	var flag int
	var snapshot int
	//gompher atomic read
	snapshot = flag
	_ = snapshot
}
`
	got := runTransform(t, src)
	if !strings.Contains(got, "snapshot = runtime.AtomicLoadInt(&flag)") {
		t.Errorf("expected snapshot = AtomicLoadInt(&flag), got:\n%s", got)
	}
}

// TestTransform_Atomic_AddsRuntimeImport verifies the runtime import is
// injected for an atomic directive in a file with no prior runtime import.
func TestTransform_Atomic_AddsRuntimeImport(t *testing.T) {
	src := `package main

func main() {
	var c int
	//gompher atomic update
	c++
	_ = c
}
`
	got := runTransform(t, src)
	if !strings.Contains(got, `"github.com/gomphermp/gomphermp/pkg/runtime"`) {
		t.Errorf("expected runtime import, got:\n%s", got)
	}
}

// ---------------------------------------------------------------------------
// Error paths reachable from real source
// ---------------------------------------------------------------------------

// TestTransform_Atomic_UnsupportedOperator verifies that a multiplicative
// op-assignment (which sync/atomic cannot express) is rejected with a clear
// error rather than producing wrong code.
func TestTransform_Atomic_UnsupportedOperator(t *testing.T) {
	src := `package main

func main() {
	var x int
	//gompher atomic update
	x *= 2
	_ = x
}
`
	err := transformErr(t, src)
	if err == nil || !strings.Contains(err.Error(), "unsupported operator") {
		t.Errorf("expected unsupported operator error, got: %v", err)
	}
}

// TestTransform_Atomic_ReadNonAddressableSource verifies that an atomic read
// whose source is not an addressable variable (here a binary expression) is
// rejected, since &(a+b) cannot be taken.
func TestTransform_Atomic_ReadNonAddressableSource(t *testing.T) {
	src := `package main

func main() {
	var a, b, v int
	//gompher atomic read
	v = a + b
	_ = v
}
`
	err := transformErr(t, src)
	if err == nil || !strings.Contains(err.Error(), "addressable") {
		t.Errorf("expected addressable-source error, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Error paths requiring constructed directives
// ---------------------------------------------------------------------------

func parseEmpty(t *testing.T) *parser.ParseResult {
	t.Helper()
	parsed, err := parser.Parse("package main\n\nfunc main() {}\n")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return parsed
}

// TestTransformAtomic_UnknownMode verifies that an out-of-range mode is
// reported. The parser only produces read/write/update, so this defends the
// transformer against an unexpected value.
func TestTransformAtomic_UnknownMode(t *testing.T) {
	err := transformAtomic(parseEmpty(t), parser.AtomicDirective{Mode: "bogus"})
	if err == nil || !strings.Contains(err.Error(), "unknown mode") {
		t.Errorf("expected unknown mode error, got: %v", err)
	}
}

// TestTransformAtomic_UpdateWrongNodeType verifies the update path rejects a
// node that is neither IncDecStmt nor AssignStmt.
func TestTransformAtomic_UpdateWrongNodeType(t *testing.T) {
	err := transformAtomic(parseEmpty(t), parser.AtomicDirective{
		Mode: "update",
		Node: &ast.ExprStmt{X: &ast.Ident{Name: "x"}},
	})
	if err == nil || !strings.Contains(err.Error(), "expected x++") {
		t.Errorf("expected wrong-node-type error, got: %v", err)
	}
}

// TestTransformAtomic_UpdateMultiTarget verifies that a multi-target
// op-assignment is rejected.
func TestTransformAtomic_UpdateMultiTarget(t *testing.T) {
	node := &ast.AssignStmt{
		Lhs: []ast.Expr{&ast.Ident{Name: "a"}, &ast.Ident{Name: "b"}},
		Tok: token.ADD_ASSIGN,
		Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "1"}},
	}
	err := transformAtomic(parseEmpty(t), parser.AtomicDirective{Mode: "update", Node: node})
	if err == nil || !strings.Contains(err.Error(), "single-target") {
		t.Errorf("expected single-target error, got: %v", err)
	}
}

// TestTransformAtomic_UpdateIncDecNotInAST verifies the replaceStmt failure
// branch for an IncDecStmt that is not reachable from the file AST.
func TestTransformAtomic_UpdateIncDecNotInAST(t *testing.T) {
	node := &ast.IncDecStmt{X: &ast.Ident{Name: "x"}, Tok: token.INC}
	err := transformAtomic(parseEmpty(t), parser.AtomicDirective{Mode: "update", Node: node})
	if err == nil || !strings.Contains(err.Error(), "not found in AST") {
		t.Errorf("expected not-found-in-AST error, got: %v", err)
	}
}

// TestTransformAtomic_UpdateAssignNotInAST verifies the replaceStmt failure
// branch for an op-assignment that is not reachable from the file AST.
func TestTransformAtomic_UpdateAssignNotInAST(t *testing.T) {
	node := &ast.AssignStmt{
		Lhs: []ast.Expr{&ast.Ident{Name: "x"}},
		Tok: token.ADD_ASSIGN,
		Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "1"}},
	}
	err := transformAtomic(parseEmpty(t), parser.AtomicDirective{Mode: "update", Node: node})
	if err == nil || !strings.Contains(err.Error(), "not found in AST") {
		t.Errorf("expected not-found-in-AST error, got: %v", err)
	}
}

// TestTransformAtomic_WriteWrongNodeType verifies the write path rejects a
// non-assignment node.
func TestTransformAtomic_WriteWrongNodeType(t *testing.T) {
	err := transformAtomic(parseEmpty(t), parser.AtomicDirective{
		Mode: "write",
		Node: &ast.IncDecStmt{X: &ast.Ident{Name: "x"}, Tok: token.INC},
	})
	if err == nil || !strings.Contains(err.Error(), "expected an assignment") {
		t.Errorf("expected assignment error, got: %v", err)
	}
}

// TestTransformAtomic_WriteNonSimple verifies the write path rejects an
// assignment that is not the simple x = e form (here a := define).
func TestTransformAtomic_WriteNonSimple(t *testing.T) {
	node := &ast.AssignStmt{
		Lhs: []ast.Expr{&ast.Ident{Name: "x"}},
		Tok: token.DEFINE, // := rather than =
		Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "1"}},
	}
	err := transformAtomic(parseEmpty(t), parser.AtomicDirective{Mode: "write", Node: node})
	if err == nil || !strings.Contains(err.Error(), "simple x = e") {
		t.Errorf("expected simple-assignment error, got: %v", err)
	}
}

// TestTransformAtomic_WriteNotInAST verifies the replaceStmt failure branch for
// a write whose assignment is not reachable from the AST.
func TestTransformAtomic_WriteNotInAST(t *testing.T) {
	node := &ast.AssignStmt{
		Lhs: []ast.Expr{&ast.Ident{Name: "x"}},
		Tok: token.ASSIGN,
		Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "1"}},
	}
	err := transformAtomic(parseEmpty(t), parser.AtomicDirective{Mode: "write", Node: node})
	if err == nil || !strings.Contains(err.Error(), "not found in AST") {
		t.Errorf("expected not-found-in-AST error, got: %v", err)
	}
}

// TestTransformAtomic_ReadWrongNodeType verifies the read path rejects a
// non-assignment node.
func TestTransformAtomic_ReadWrongNodeType(t *testing.T) {
	err := transformAtomic(parseEmpty(t), parser.AtomicDirective{
		Mode: "read",
		Node: &ast.IncDecStmt{X: &ast.Ident{Name: "x"}, Tok: token.INC},
	})
	if err == nil || !strings.Contains(err.Error(), "expected an assignment") {
		t.Errorf("expected assignment error, got: %v", err)
	}
}

// TestTransformAtomic_ReadNonSimple verifies the read path rejects a := define.
func TestTransformAtomic_ReadNonSimple(t *testing.T) {
	node := &ast.AssignStmt{
		Lhs: []ast.Expr{&ast.Ident{Name: "v"}},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{&ast.Ident{Name: "x"}},
	}
	err := transformAtomic(parseEmpty(t), parser.AtomicDirective{Mode: "read", Node: node})
	if err == nil || !strings.Contains(err.Error(), "simple v = x") {
		t.Errorf("expected simple-assignment error, got: %v", err)
	}
}

// TestTransform_PropagatesAtomicError verifies that a failing transformAtomic
// aborts Transform and returns nil.
func TestTransform_PropagatesAtomicError(t *testing.T) {
	parsed := parseEmpty(t)
	parsed.Nodes = append(parsed.Nodes, parser.AnnotatedNode{
		Directive: parser.AtomicDirective{Mode: "bogus"},
	})
	got, err := Transform(parsed)
	if err == nil {
		t.Fatal("expected propagated error")
	}
	if got != nil {
		t.Errorf("expected nil result on error, got %v", got)
	}
}

// ---------------------------------------------------------------------------
// isAddressable
// ---------------------------------------------------------------------------

// TestIsAddressable covers each recognized addressable form and the rejection
// of a non-addressable expression.
func TestIsAddressable(t *testing.T) {
	addressable := []ast.Expr{
		&ast.Ident{Name: "x"},
		&ast.SelectorExpr{X: &ast.Ident{Name: "s"}, Sel: &ast.Ident{Name: "f"}},
		&ast.IndexExpr{X: &ast.Ident{Name: "a"}, Index: &ast.BasicLit{Kind: token.INT, Value: "0"}},
	}
	for _, e := range addressable {
		if !isAddressable(e) {
			t.Errorf("expected %T to be addressable", e)
		}
	}

	notAddressable := []ast.Expr{
		&ast.BasicLit{Kind: token.INT, Value: "1"},
		&ast.BinaryExpr{X: &ast.Ident{Name: "a"}, Op: token.ADD, Y: &ast.Ident{Name: "b"}},
		&ast.CallExpr{Fun: &ast.Ident{Name: "f"}},
	}
	for _, e := range notAddressable {
		if isAddressable(e) {
			t.Errorf("expected %T to be non-addressable", e)
		}
	}
}
