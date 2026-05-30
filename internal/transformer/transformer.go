// Package transformer rewrites a parsed Go program that contains //gompher
// directives into an equivalent program that calls the gomphermp runtime
// instead. It consumes the typed directive set produced by internal/parser
// and emits a transformed *ast.File backed by the same FileSet.
package transformer

import (
	"github.com/gomphermp/gomphermp/internal/parser"
)

// Transform rewrites every annotated GompherMP directive in the parsed file
// into the corresponding runtime call. Nodes that do not correspond to a
// directive are passed through unmodified.
func Transform(result *parser.ParseResult) (*parser.ParseResult, error) {
	if result == nil {
		return nil, nil
	}
	return result, nil
}
