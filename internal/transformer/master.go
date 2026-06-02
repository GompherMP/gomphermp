package transformer

import (
	"go/ast"

	"github.com/gomphermp/gomphermp/internal/parser"
)

// threadIDParamName is the canonical parameter name the transformer emits
// for the per-thread identifier of Parallel and ParallelFor regions. Master
// directives reference this name when calling runtime.Master, so it must
// stay in sync with whatever buildClosureWithIntParam emits for those
// directives (Phase 3 onward).
const threadIDParamName = "threadID"

// transformMaster rewrites a //gompher master directive into a
// runtime.Master(threadID, func() { body }) call. The threadID identifier
// must be visible in the enclosing scope — typically the parameter of the
// surrounding runtime.Parallel(func(threadID int) {...}) call. When master
// appears outside a parallel region the emitted code will fail to compile
// with "undefined: threadID", which surfaces the semantic error cleanly via
// the Go compiler instead of needing a separate validation pass.
func transformMaster(file *parser.ParseResult, d parser.MasterDirective) error {
	return transformBlockDirective(
		file, d.Node, d.Pos, d.Line,
		"Master",
		&ast.Ident{Name: threadIDParamName},
	)
}
