package transformer

import (
	"fmt"

	"github.com/gomphermp/gomphermp/internal/parser"
)

// transformBarrier rewrites a //gompher barrier directive into a
// runtime.Barrier() call. Unlike the block and loop directives, barrier has no
// associated AST node — it is a standalone synchronization point — so there is
// nothing to replace. Instead the call is spliced into the enclosing block at
// the directive's position, between the statements that precede and follow it.
//
// The insertion is positional (insertCallAtPos locates the innermost block
// whose braces straddle the directive). This keeps working even after an
// enclosing //gompher parallel has already been rewritten into a closure,
// because the parallel transform reuses the original body block — its brace
// positions, and therefore the barrier's containing block, are unchanged.
func transformBarrier(result *parser.ParseResult, d parser.BarrierDirective) error {
	call := buildRuntimeCall("Barrier")
	if !insertCallAtPos(result.File, d.Pos, call) {
		return fmt.Errorf("Barrier at line %d: could not find containing block", d.Line)
	}
	removeDirectiveComment(result.File, d.Pos)
	return nil
}
