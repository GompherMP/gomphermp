package transformer

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/gomphermp/gomphermp/internal/parser"
)

// transformAtomic rewrites a //gompher atomic directive into a lock-free call
// against the runtime's int atomics. The mode selects the operation.
func transformAtomic(result *parser.ParseResult, d parser.AtomicDirective) error {
	mode := d.Mode
	if mode == "" {
		mode = "update"
	}

	switch mode {
	case "update":
		return transformAtomicUpdate(result, d)
	case "write":
		return transformAtomicWrite(result, d)
	case "read":
		return transformAtomicRead(result, d)
	default:
		return fmt.Errorf("atomic at line %d: unknown mode %q", d.Line, d.Mode)
	}
}

// transformAtomicUpdate handles the update forms. The target may be an
// IncDecStmt (x++, x--) or an additive op-assignment (x += e, x -= e); both
// become AtomicAddInt with the appropriate delta.
func transformAtomicUpdate(result *parser.ParseResult, d parser.AtomicDirective) error {
	switch node := d.Node.(type) {
	case *ast.IncDecStmt:
		delta := "1"
		if node.Tok == token.DEC {
			delta = "-1"
		}
		call := buildRuntimeCall("AtomicAddInt",
			buildAddrOf(node.X),
			&ast.BasicLit{Kind: token.INT, Value: delta},
		)
		if !replaceStmt(result.File, node, call) {
			return fmt.Errorf("atomic update at line %d: statement not found in AST", d.Line)
		}

	case *ast.AssignStmt:
		if len(node.Lhs) != 1 || len(node.Rhs) != 1 {
			return fmt.Errorf("atomic update at line %d: expected single-target assignment", d.Line)
		}
		var delta ast.Expr
		switch node.Tok {
		case token.ADD_ASSIGN:
			delta = node.Rhs[0]
		case token.SUB_ASSIGN:
			// x -= e becomes AtomicAddInt(&x, -(e)); parenthesize so a compound
			// right-hand side keeps its meaning under the unary minus.
			delta = &ast.UnaryExpr{Op: token.SUB, X: &ast.ParenExpr{X: node.Rhs[0]}}
		default:
			return fmt.Errorf("atomic update at line %d: unsupported operator %s (only += and -= are atomic)", d.Line, node.Tok)
		}
		call := buildRuntimeCall("AtomicAddInt", buildAddrOf(node.Lhs[0]), delta)
		if !replaceStmt(result.File, node, call) {
			return fmt.Errorf("atomic update at line %d: statement not found in AST", d.Line)
		}

	default:
		return fmt.Errorf("atomic update at line %d: expected x++/x--/x+=e/x-=e, got %T", d.Line, d.Node)
	}

	removeDirectiveComment(result.File, d.Pos)
	return nil
}

// transformAtomicWrite handles x = e, replacing the assignment with an atomic
// store of e into x.
func transformAtomicWrite(result *parser.ParseResult, d parser.AtomicDirective) error {
	assign, ok := d.Node.(*ast.AssignStmt)
	if !ok {
		return fmt.Errorf("atomic write at line %d: expected an assignment, got %T", d.Line, d.Node)
	}
	if assign.Tok != token.ASSIGN || len(assign.Lhs) != 1 || len(assign.Rhs) != 1 {
		return fmt.Errorf("atomic write at line %d: expected a simple x = e assignment", d.Line)
	}

	call := buildRuntimeCall("AtomicStoreInt", buildAddrOf(assign.Lhs[0]), assign.Rhs[0])
	if !replaceStmt(result.File, assign, call) {
		return fmt.Errorf("atomic write at line %d: statement not found in AST", d.Line)
	}

	removeDirectiveComment(result.File, d.Pos)
	return nil
}

// transformAtomicRead handles v = x, where the atomic variable is the source
// on the right-hand side. The assignment statement is kept; only its RHS is
// replaced with an atomic load, so the surrounding code (the destination v) is
// untouched.
func transformAtomicRead(result *parser.ParseResult, d parser.AtomicDirective) error {
	assign, ok := d.Node.(*ast.AssignStmt)
	if !ok {
		return fmt.Errorf("atomic read at line %d: expected an assignment, got %T", d.Line, d.Node)
	}
	if assign.Tok != token.ASSIGN || len(assign.Rhs) != 1 {
		return fmt.Errorf("atomic read at line %d: expected a simple v = x assignment", d.Line)
	}
	if !isAddressable(assign.Rhs[0]) {
		return fmt.Errorf("atomic read at line %d: source must be an addressable variable", d.Line)
	}

	assign.Rhs[0] = buildRuntimeCallExpr("AtomicLoadInt", buildAddrOf(assign.Rhs[0]))
	removeDirectiveComment(result.File, d.Pos)
	return nil
}

// isAddressable reports whether e is one of the expression forms whose address
// can be taken (a plain variable, a struct field, or an indexed element). It
// guards the atomic read rewrite from generating &(non-addressable), which
// would not compile.
func isAddressable(e ast.Expr) bool {
	switch e.(type) {
	case *ast.Ident, *ast.SelectorExpr, *ast.IndexExpr:
		return true
	default:
		return false
	}
}
