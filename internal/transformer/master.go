package transformer

import (
	"go/ast"

	"github.com/gomphermp/gomphermp/internal/parser"
)

// threadIDParamName is the canonical parameter name the transformer emits
// for the per-thread identifier of Parallel and ParallelFor regions. Master
// directives reference this name when calling runtime.Master.
const threadIDParamName = "threadID"

// transformMaster rewrites a //gompher master directive into a
// runtime.Master(threadID, func() { body }) call. The threadID identifier
// must be visible in the enclosing scope. When master appears outside a
// parallel region the emitted code will fail to compile.
func transformMaster(file *parser.ParseResult, d parser.MasterDirective) error {
	return transformBlockDirective(
		file, d.Node, d.Pos, d.Line,
		"Master",
		&ast.Ident{Name: threadIDParamName},
	)
}
