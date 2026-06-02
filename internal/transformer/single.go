package transformer

import "github.com/gomphermp/gomphermp/internal/parser"

// transformSingle rewrites a //gompher single directive into a
// runtime.Single(func() { body }) call. The runtime guarantees that the
// body executes exactly once across the team.
func transformSingle(file *parser.ParseResult, d parser.SingleDirective) error {
	return transformBlockDirective(
		file, d.Node, d.Pos, d.Line,
		"Single",
	)
}
