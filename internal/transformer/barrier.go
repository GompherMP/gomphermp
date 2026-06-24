package transformer

import (
	"fmt"

	"github.com/gomphermp/gomphermp/internal/parser"
)

// transformBarrier rewrites a //gompher barrier directive into a
// runtime.Barrier() call.
func transformBarrier(result *parser.ParseResult, d parser.BarrierDirective) error {
	call := buildRuntimeCall("Barrier")
	if !insertCallAtPos(result.File, d.Pos, call) {
		return fmt.Errorf("Barrier at line %d: could not find containing block", d.Line)
	}
	removeDirectiveComment(result.File, d.Pos)
	return nil
}
