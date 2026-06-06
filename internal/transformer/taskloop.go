package transformer

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/gomphermp/gomphermp/internal/parser"
)

// transformTaskloop rewrites a //gompher taskloop directive into
// runtime.Taskloop(func(i int) { body }, iterations, grainsize).
// The directive must annotate a canonical for loop: for i := 0; i < N; i++.
func transformTaskloop(result *parser.ParseResult, d parser.TaskloopDirective) error {
	forStmt, ok := d.Node.(*ast.ForStmt)
	if !ok {
		return fmt.Errorf("Taskloop at line %d: expected *ast.ForStmt, got %T", d.Line, d.Node)
	}

	loopVar, err := extractLoopVar(forStmt)
	if err != nil {
		return fmt.Errorf("Taskloop at line %d: %w", d.Line, err)
	}

	bound, err := extractUpperBound(forStmt)
	if err != nil {
		return fmt.Errorf("Taskloop at line %d: %w", d.Line, err)
	}

	grainsize := "1"
	for _, clause := range d.Clauses {
		if gs, ok := clause.(parser.GrainsizeClause); ok {
			grainsize = gs.Size
			break
		}
	}

	closure := buildClosureWithIntParam(forStmt.Body, loopVar)
	grainsizeLit := &ast.BasicLit{Kind: token.INT, Value: grainsize}
	call := buildRuntimeCall("Taskloop", closure, bound, grainsizeLit)

	if !replaceForStmt(result.File, forStmt, call) {
		return fmt.Errorf("Taskloop at line %d: for statement not found in AST", d.Line)
	}
	removeDirectiveComment(result.File, d.Pos)

	return nil
}

// extractLoopVar returns the name of the iteration variable from a for loop's
// init statement. Requires the canonical form: i := 0.
func extractLoopVar(forStmt *ast.ForStmt) (string, error) {
	if forStmt.Init == nil {
		return "", fmt.Errorf("for loop has no init statement")
	}
	assign, ok := forStmt.Init.(*ast.AssignStmt)
	if !ok || len(assign.Lhs) != 1 {
		return "", fmt.Errorf("expected simple assignment in for init, got %T", forStmt.Init)
	}
	ident, ok := assign.Lhs[0].(*ast.Ident)
	if !ok {
		return "", fmt.Errorf("expected identifier on left side of for init")
	}
	return ident.Name, nil
}

// extractUpperBound returns the upper bound expression from a for loop's
// condition. Requires the canonical form: i < N.
func extractUpperBound(forStmt *ast.ForStmt) (ast.Expr, error) {
	if forStmt.Cond == nil {
		return nil, fmt.Errorf("for loop has no condition")
	}
	binExpr, ok := forStmt.Cond.(*ast.BinaryExpr)
	if !ok {
		return nil, fmt.Errorf("expected binary expression in for condition, got %T", forStmt.Cond)
	}
	if binExpr.Op != token.LSS {
		return nil, fmt.Errorf("expected '<' operator in for condition, got %s", binExpr.Op)
	}
	return binExpr.Y, nil
}

// replaceForStmt walks file's AST looking for target *ast.ForStmt in a parent
// BlockStmt's List and substitutes it with replacement. Returns true if found.
func replaceForStmt(file *ast.File, target *ast.ForStmt, replacement ast.Stmt) bool {
	var replaced bool
	ast.Inspect(file, func(n ast.Node) bool {
		if replaced {
			return false
		}
		block, ok := n.(*ast.BlockStmt)
		if !ok {
			return true
		}
		for i, stmt := range block.List {
			if fs, ok := stmt.(*ast.ForStmt); ok && fs == target {
				block.List[i] = replacement
				replaced = true
				return false
			}
		}
		return true
	})
	return replaced
}
