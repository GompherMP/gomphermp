package transformer

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/gomphermp/gomphermp/internal/parser"
)

// transformTaskloop rewrites a //gompher taskloop directive into
// runtime.Taskloop(func(i int) { body }, iterations, grainsize).
// The directive must annotate a canonical for loop (see analyzeLoop).
//
// private/firstprivate clauses are applied per iteration: the matching
// declarations are prepended inside the loop closure, and firstprivate
// captures the outer value once before the Taskloop call.
func transformTaskloop(result *parser.ParseResult, d parser.TaskloopDirective) error {
	forStmt, ok := d.Node.(*ast.ForStmt)
	if !ok {
		return fmt.Errorf("Taskloop at line %d: expected *ast.ForStmt, got %T", d.Line, d.Node)
	}

	form, err := analyzeLoop(forStmt)
	if err != nil {
		return fmt.Errorf("Taskloop at line %d: %w", d.Line, err)
	}

	closurePrefix, capturePrefix, err := dataClausePrefixes(result, d.Clauses, d.Pos)
	if err != nil {
		return fmt.Errorf("Taskloop at line %d: %w", d.Line, err)
	}

	// The runtime drives the body with a 0-based counter over [0, count), so we
	// assemble the closure body in order:
	// - Induction recovery (only when the loop isn't the simple 0..N form):
	//   `v := lb + counter*step` reconstructs the user's loop variable first,
	//   so the rest of the body can use it.
	// - Clause prefixes (private declarations, firstprivate shadows): fresh
	//   locals that hide the outer variables for the rest of the body (placed
	//   after the recovery so the loop variable is already in scope).
	// - The user's original loop body.
	count := form.iterationCount()
	counterName := form.loopVar
	var bodyStmts []ast.Stmt
	if !form.simple() {
		counterName = loopCounterName
		bodyStmts = append(bodyStmts, form.inductionRecovery())
	}
	bodyStmts = append(bodyStmts, closurePrefix...)
	bodyStmts = append(bodyStmts, forStmt.Body.List...)

	grainsize := "1"
	for _, clause := range d.Clauses {
		if gs, ok := clause.(parser.GrainsizeClause); ok {
			grainsize = gs.Size
			break
		}
	}

	closure := buildClosureWithIntParam(&ast.BlockStmt{List: bodyStmts}, counterName)
	grainsizeLit := &ast.BasicLit{Kind: token.INT, Value: grainsize}
	call := buildRuntimeCall("Taskloop", closure, count, grainsizeLit)

	if !replaceStmtWithPrefix(result.File, forStmt, call, capturePrefix) {
		return fmt.Errorf("Taskloop at line %d: for statement not found in AST", d.Line)
	}
	removeDirectiveComment(result.File, d.Pos)

	return nil
}
