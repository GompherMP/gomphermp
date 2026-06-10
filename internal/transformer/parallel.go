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
func transformParallel(result *parser.ParseResult, d parser.ParallelDirective) error {
	body, ok := d.Node.(*ast.BlockStmt)
	if !ok {
		return fmt.Errorf("Parallel at line %d: expected *ast.BlockStmt, got %T", d.Line, d.Node)
	}

	closure := buildClosureWithIntParam(body, threadIDParamName)
	call := buildRuntimeCall("Parallel", closure)

	if !replaceBlockStmt(result.File, body, call) {
		return fmt.Errorf("Parallel at line %d: body block not found in AST", d.Line)
	}
	removeDirectiveComment(result.File, d.Pos)

	return nil
}
