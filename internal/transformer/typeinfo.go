package transformer

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
)

// resolveVarType type-checks file and returns the declared type of varName (the
// most recent definition before beforePos) as an AST expression suitable for a
// `var x T` declaration. It resolves types inferred from `:=`, from function-call
// results, and from compound expressions.
//
// Type-checking is best-effort: the enriched, not-yet-transformed source can
// contain references the checker cannot resolve (a bare threadID, for example),
// so type errors are ignored and the function resolves whatever the checker
// could type. It fails only when varName itself could not be typed.
func resolveVarType(fset *token.FileSet, file *ast.File, varName string, beforePos token.Pos) (ast.Expr, error) {
	info := &types.Info{Defs: make(map[*ast.Ident]types.Object)}
	conf := types.Config{
		Importer: importer.Default(),
		Error:    func(error) {}, // tolerate type errors; resolve what we can
	}
	// The aggregate result is intentionally ignored; the Error hook above
	// already swallowed individual errors and info.Defs is populated for
	// whatever type-checked successfully.
	_, _ = conf.Check(file.Name.Name, fset, []*ast.File{file}, info)

	var best types.Object
	var bestPos token.Pos
	for ident, obj := range info.Defs {
		if obj == nil || ident.Name != varName || ident.Pos() >= beforePos {
			continue
		}
		if ident.Pos() > bestPos {
			best, bestPos = obj, ident.Pos()
		}
	}
	if best == nil {
		return nil, fmt.Errorf("cannot resolve the type of %q (is it declared before the directive?)", varName)
	}

	// Render the resolved type as source and parse it back into an AST
	// expression. types.TypeString handles every type uniformly (int, []int,
	// map[string]int, named and qualified types). Package-qualified names use
	// the imported package's short name, which the user's file already imports.
	typeStr := types.TypeString(best.Type(), func(p *types.Package) string { return p.Name() })
	expr, err := parser.ParseExpr(typeStr)
	if err != nil {
		return nil, fmt.Errorf("cannot express type %q of %q as a declaration: %w", typeStr, varName, err)
	}
	return expr, nil
}
