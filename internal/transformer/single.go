package transformer

import (
	"fmt"
	"go/ast"

	"github.com/gomphermp/gomphermp/internal/parser"
)

// transformSingle rewrites a //gompher single directive into runtime.Single(
// func() { body }), which the runtime runs exactly once across the team.
// private/firstprivate are applied as for parallel: the shadow declarations go
// inside the closure (firstprivate also captures the outer value before the
// call). Since the body runs on one goroutine, a private here is just a fresh
// local shadowing the outer variable.
func transformSingle(file *parser.ParseResult, d parser.SingleDirective) error {
	body, ok := d.Node.(*ast.BlockStmt)
	if !ok {
		return fmt.Errorf("Single at line %d: expected *ast.BlockStmt, got %T", d.Line, d.Node)
	}

	closurePrefix, capturePrefix, err := dataClausePrefixes(file, d.Clauses, d.Pos)
	if err != nil {
		return fmt.Errorf("Single at line %d: %w", d.Line, err)
	}

	closure := buildClosure(wrapBodyWithPrefix(body, closurePrefix))
	call := buildRuntimeCall("Single", closure)

	var replaced bool
	if len(capturePrefix) > 0 {
		replaced = replaceBlockStmtWithPrefix(file.File, body, call, capturePrefix)
	} else {
		replaced = replaceBlockStmt(file.File, body, call)
	}
	if !replaced {
		return fmt.Errorf("Single at line %d: body block not found in AST", d.Line)
	}
	removeDirectiveComment(file.File, d.Pos)

	return nil
}
