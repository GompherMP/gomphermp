package transformer

import (
	"go/ast"
	goparser "go/parser"
	"go/token"
	"strings"
	"testing"
)

// parseLoop parses a single for statement from source and returns its AST node.
func parseLoop(t *testing.T, loop string) *ast.ForStmt {
	t.Helper()
	src := "package p\nfunc f() {\n\tN := 0\n\tj := 0\n\t_, _ = N, j\n\t" + loop + " {\n\t\t_ = i\n\t}\n}\n"
	file, err := goparser.ParseFile(token.NewFileSet(), "", src, 0)
	if err != nil {
		t.Fatalf("parse %q: %v", loop, err)
	}
	var found *ast.ForStmt
	ast.Inspect(file, func(n ast.Node) bool {
		if fs, ok := n.(*ast.ForStmt); ok && found == nil {
			found = fs
			return false
		}
		return true
	})
	if found == nil {
		t.Fatalf("no for statement found in %q", loop)
	}
	return found
}

// TestAnalyzeLoop_AcceptedForms verifies the canonical forms analyzeLoop accepts
// and the direction/step/bound it decodes for each.
func TestAnalyzeLoop_AcceptedForms(t *testing.T) {
	cases := []struct {
		loop       string
		descending bool
		simple     bool
		relop      token.Token
	}{
		{"for i := 0; i < N; i++", false, true, token.LSS},
		{"for i := 0; i < N; i += 1", false, true, token.LSS},
		{"for i := 5; i < N; i++", false, false, token.LSS},
		{"for i := 0; i <= N; i++", false, false, token.LEQ},
		{"for i := 0; i < N; i += 2", false, false, token.LSS},
		{"for i := N; i > 0; i--", true, false, token.GTR},
		{"for i := N; i >= 0; i -= 3", true, false, token.GEQ},
	}
	for _, tc := range cases {
		t.Run(tc.loop, func(t *testing.T) {
			f, err := analyzeLoop(parseLoop(t, tc.loop))
			if err != nil {
				t.Fatalf("expected acceptance, got: %v", err)
			}
			if f.descending != tc.descending {
				t.Errorf("descending = %v, want %v", f.descending, tc.descending)
			}
			if f.simple() != tc.simple {
				t.Errorf("simple = %v, want %v", f.simple(), tc.simple)
			}
			if f.relop != tc.relop {
				t.Errorf("relop = %v, want %v", f.relop, tc.relop)
			}
			if f.loopVar != "i" {
				t.Errorf("loopVar = %q, want i", f.loopVar)
			}
		})
	}
}

// TestAnalyzeLoop_RejectedForms verifies analyzeLoop rejects non-canonical loops
// with a diagnostic mentioning the offending part.
func TestAnalyzeLoop_RejectedForms(t *testing.T) {
	cases := []struct {
		loop    string
		wantSub string
	}{
		{"for i := 0; j < N; i++", "condition must test"},
		{"for i := 0; i < N; i *= 2", "step must be"},
		{"for i := 0; i < N; i--", "direction disagree"},
		{"for i := 0; i > N; i++", "direction disagree"},
		{"for i := 0; i != N; i++", "not supported"},
	}
	for _, tc := range cases {
		t.Run(tc.loop, func(t *testing.T) {
			_, err := analyzeLoop(parseLoop(t, tc.loop))
			if err == nil {
				t.Fatalf("expected rejection for %q", tc.loop)
			}
			if !strings.Contains(err.Error(), tc.wantSub) {
				t.Errorf("expected error containing %q, got: %v", tc.wantSub, err)
			}
		})
	}
}

// TestAtomize verifies the parenthesization helper wraps expressions that would
// re-associate under a surrounding * or / and leaves atomic operands bare.
func TestAtomize(t *testing.T) {
	atomic := []ast.Expr{
		&ast.Ident{Name: "k"},
		intLit("2"),
		&ast.CallExpr{Fun: &ast.Ident{Name: "len"}, Args: []ast.Expr{&ast.Ident{Name: "x"}}},
	}
	for _, e := range atomic {
		if _, wrapped := atomize(e).(*ast.ParenExpr); wrapped {
			t.Errorf("atomize(%T) should not wrap an atomic operand", e)
		}
	}
	compound := []ast.Expr{
		&ast.BinaryExpr{X: &ast.Ident{Name: "k"}, Op: token.MUL, Y: intLit("2")},
		&ast.UnaryExpr{Op: token.SUB, X: &ast.Ident{Name: "k"}},
		&ast.StarExpr{X: &ast.Ident{Name: "p"}},
	}
	for _, e := range compound {
		if _, wrapped := atomize(e).(*ast.ParenExpr); !wrapped {
			t.Errorf("atomize(%T) should wrap a re-associating operand", e)
		}
	}
}

// TestAnalyzeLoop_StructuralRejections covers early structural guards that a
// parsed loop cannot reach (built from synthetic nodes): missing init, a
// non-assignment init, and a missing condition.
func TestAnalyzeLoop_StructuralRejections(t *testing.T) {
	cases := []struct {
		name    string
		loop    *ast.ForStmt
		wantSub string
	}{
		{
			name:    "no init",
			loop:    &ast.ForStmt{Cond: &ast.BinaryExpr{X: &ast.Ident{Name: "i"}, Op: token.LSS, Y: &ast.Ident{Name: "n"}}, Post: &ast.IncDecStmt{X: &ast.Ident{Name: "i"}, Tok: token.INC}, Body: &ast.BlockStmt{}},
			wantSub: "missing init",
		},
		{
			name:    "init not assignment",
			loop:    &ast.ForStmt{Init: &ast.ExprStmt{X: &ast.Ident{Name: "x"}}, Body: &ast.BlockStmt{}},
			wantSub: "init must be",
		},
		{
			name: "no condition",
			loop: &ast.ForStmt{
				Init: &ast.AssignStmt{Tok: token.DEFINE, Lhs: []ast.Expr{&ast.Ident{Name: "i"}}, Rhs: []ast.Expr{intLit("0")}},
				Body: &ast.BlockStmt{},
			},
			wantSub: "missing condition",
		},
		{
			name:    "init target not identifier",
			loop:    &ast.ForStmt{Init: &ast.AssignStmt{Tok: token.DEFINE, Lhs: []ast.Expr{intLit("0")}, Rhs: []ast.Expr{intLit("0")}}, Body: &ast.BlockStmt{}},
			wantSub: "left side must be an identifier",
		},
		{
			name: "incdec step not on loop var",
			loop: &ast.ForStmt{
				Init: &ast.AssignStmt{Tok: token.DEFINE, Lhs: []ast.Expr{&ast.Ident{Name: "i"}}, Rhs: []ast.Expr{intLit("0")}},
				Cond: &ast.BinaryExpr{X: &ast.Ident{Name: "i"}, Op: token.LSS, Y: &ast.Ident{Name: "n"}},
				Post: &ast.IncDecStmt{X: intLit("0"), Tok: token.INC},
				Body: &ast.BlockStmt{},
			},
			wantSub: "step must advance",
		},
		{
			name: "assign step multiple targets",
			loop: &ast.ForStmt{
				Init: &ast.AssignStmt{Tok: token.DEFINE, Lhs: []ast.Expr{&ast.Ident{Name: "i"}}, Rhs: []ast.Expr{intLit("0")}},
				Cond: &ast.BinaryExpr{X: &ast.Ident{Name: "i"}, Op: token.LSS, Y: &ast.Ident{Name: "n"}},
				Post: &ast.AssignStmt{Tok: token.ADD_ASSIGN, Lhs: []ast.Expr{&ast.Ident{Name: "i"}, &ast.Ident{Name: "j"}}, Rhs: []ast.Expr{intLit("1"), intLit("1")}},
				Body: &ast.BlockStmt{},
			},
			wantSub: "step must be",
		},
		{
			name: "assign step target not identifier",
			loop: &ast.ForStmt{
				Init: &ast.AssignStmt{Tok: token.DEFINE, Lhs: []ast.Expr{&ast.Ident{Name: "i"}}, Rhs: []ast.Expr{intLit("0")}},
				Cond: &ast.BinaryExpr{X: &ast.Ident{Name: "i"}, Op: token.LSS, Y: &ast.Ident{Name: "n"}},
				Post: &ast.AssignStmt{Tok: token.ADD_ASSIGN, Lhs: []ast.Expr{intLit("0")}, Rhs: []ast.Expr{intLit("1")}},
				Body: &ast.BlockStmt{},
			},
			wantSub: "step must advance",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := analyzeLoop(tc.loop)
			if err == nil {
				t.Fatalf("expected rejection")
			}
			if !strings.Contains(err.Error(), tc.wantSub) {
				t.Errorf("expected error containing %q, got: %v", tc.wantSub, err)
			}
		})
	}
}
