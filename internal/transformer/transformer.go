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
//
// Dispatch happens via a type switch over the concrete directive carried by
// each AnnotatedNode. Each directive's handler mutates the AST in place. When
// at least one directive is rewritten, the gomphermp runtime import is added
// so the emitted calls compile.
func Transform(result *parser.ParseResult) (*parser.ParseResult, error) {
	if result == nil {
		return nil, nil
	}

	var emittedRuntime bool
	for _, node := range result.Nodes {
		switch d := node.Directive.(type) {
		case parser.CriticalDirective:
			if err := transformCritical(result, d); err != nil {
				return nil, err
			}
			emittedRuntime = true
		}
	}

	if emittedRuntime {
		ensureRuntimeImport(result.File)
	}

	return result, nil
}
