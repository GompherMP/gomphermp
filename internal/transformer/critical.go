package transformer

import (
	"fmt"
	"go/ast"

	"github.com/gomphermp/gomphermp/internal/parser"
)

// transformCritical rewrites a //gompher critical directive into a
// runtime.Critical call. The directive's BlockStmt body is wrapped in a
// closure and passed as the second argument; the lock name (empty for the
// anonymous variant) is passed as the first argument.
// The closure preserves the original body unchanged, so variables captured
// from the enclosing scope continue to work without any special handling —
// Go closures capture by reference, which is exactly the semantics OpenMP
// associates with shared (default) variables inside a critical region.
func transformCritical(file *parser.ParseResult, d parser.CriticalDirective) error {
	body, ok := d.Node.(*ast.BlockStmt)
	if !ok {
		return fmt.Errorf("critical at line %d: expected *ast.BlockStmt, got %T", d.Line, d.Node)
	}

	call := buildRuntimeCall(
		"Critical",
		buildStringLit(d.Name),
		buildClosure(body),
	)

	if !replaceBlockStmt(file.File, body, call) {
		return fmt.Errorf("critical at line %d: body block not found in AST", d.Line)
	}
	removeDirectiveComment(file.File, d.Pos)

	return nil
}
