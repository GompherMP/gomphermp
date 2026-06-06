package transformer

import "github.com/gomphermp/gomphermp/internal/parser"

// transformCritical rewrites a //gompher critical directive into a
// runtime.Critical(name, func() { body }) call. The lock name (empty for
// the anonymous variant) is passed as the first argument.
func transformCritical(file *parser.ParseResult, d parser.CriticalDirective) error {
	return transformBlockDirective(
		file, d.Node, d.Pos, d.Line,
		"Critical",
		buildStringLit(d.Name),
	)
}
