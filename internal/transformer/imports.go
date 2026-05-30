package transformer

import (
	"go/ast"
	"go/token"
	"strconv"
)

// runtimeImportPath is the canonical import path of the gomphermp runtime.
// Every transformed directive except "atomic" emits a call into this package,
// so the transformer must guarantee the file imports it before such calls
// can compile.
const runtimeImportPath = "github.com/gomphermp/gomphermp/pkg/runtime"

// ensureRuntimeImport adds an import of the gomphermp runtime to file if it
// is not already imported. The operation is idempotent: invoking it multiple
// times on the same file has no additional effect.
func ensureRuntimeImport(file *ast.File) {
	quoted := strconv.Quote(runtimeImportPath)

	for _, imp := range file.Imports {
		if imp.Path != nil && imp.Path.Value == quoted {
			return
		}
	}

	newImport := &ast.ImportSpec{
		Path: &ast.BasicLit{
			Kind:  token.STRING,
			Value: quoted,
		},
	}

	var importDecl *ast.GenDecl
	for _, decl := range file.Decls {
		if gd, ok := decl.(*ast.GenDecl); ok && gd.Tok == token.IMPORT {
			importDecl = gd
			break
		}
	}

	if importDecl == nil {
		importDecl = &ast.GenDecl{
			Tok:    token.IMPORT,
			Lparen: token.NoPos,
			Specs:  []ast.Spec{newImport},
		}
		file.Decls = append([]ast.Decl{importDecl}, file.Decls...)
	} else {
		importDecl.Specs = append(importDecl.Specs, newImport)
		// Force the grouped form so multiple specs render as a block,
		// not as two adjacent single-import lines.
		if importDecl.Lparen == token.NoPos {
			importDecl.Lparen = importDecl.TokPos + 1
			importDecl.Rparen = importDecl.TokPos + 2
		}
	}

	file.Imports = append(file.Imports, newImport)
}
