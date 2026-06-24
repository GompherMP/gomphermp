package transformer

import (
	"fmt"
	"go/ast"

	"github.com/gomphermp/gomphermp/internal/parser"
)

// transformParallel rewrites a //gompher parallel directive into a
// runtime.Parallel(func(threadID int) { body }) call. The parallel closure
// receives the per-goroutine thread identifier, so the body can branch on
// threadID and so nested master directives compile.
//
// private/firstprivate clauses are applied by prepending the matching
// declarations inside the closure (and a capture before the call for
// firstprivate); shared is the default and needs nothing. Because the closure
// runs once per goroutine, a private variable here is correctly one copy per
// goroutine.
func transformParallel(result *parser.ParseResult, d parser.ParallelDirective) error {
	body, ok := d.Node.(*ast.BlockStmt)
	if !ok {
		return fmt.Errorf("Parallel at line %d: expected *ast.BlockStmt, got %T", d.Line, d.Node)
	}

	closurePrefix, capturePrefix, err := dataClausePrefixes(result, d.Clauses, d.Pos)
	if err != nil {
		return fmt.Errorf("Parallel at line %d: %w", d.Line, err)
	}

	closure := buildClosureWithIntParam(wrapBodyWithPrefix(body, closurePrefix), threadIDParamName)
	call := buildRuntimeCall("Parallel", closure)

	var replaced bool
	if len(capturePrefix) > 0 {
		replaced = replaceBlockStmtWithPrefix(result.File, body, call, capturePrefix)
	} else {
		replaced = replaceBlockStmt(result.File, body, call)
	}
	if !replaced {
		return fmt.Errorf("Parallel at line %d: body block not found in AST", d.Line)
	}
	removeDirectiveComment(result.File, d.Pos)

	return nil
}
