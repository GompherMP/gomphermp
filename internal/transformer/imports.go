package transformer

import (
	"go/ast"
	"go/token"
	"strconv"
)

const (
	// runtimeImportPath is the import path of the gomphermp runtime that
	// transformed programs are linked against.
	runtimeImportPath = "github.com/gomphermp/gomphermp/pkg/runtime"

	// unsafeImportPath is needed by the depend rewrite, which passes variable
	// addresses as uintptr via unsafe.Pointer.
	unsafeImportPath = "unsafe"
)

// ensureRuntimeImport adds an import of the gomphermp runtime to file if it is
// not already imported.
func ensureRuntimeImport(file *ast.File) { ensureImport(file, runtimeImportPath) }

// ensureUnsafeImport adds an import of "unsafe" to file if not already present.
func ensureUnsafeImport(file *ast.File) { ensureImport(file, unsafeImportPath) }

// ensureImport adds an import of path to file if it is not already imported.
// The operation is idempotent: invoking it multiple times on the same file has
// no additional effect, which lets directive handlers call it unconditionally
// without coordinating among themselves. When the file has no existing import
// block a new one is created at the top of the declarations.
func ensureImport(file *ast.File, path string) {
	quoted := strconv.Quote(path)

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
		if importDecl.Lparen == token.NoPos {
			importDecl.Lparen = importDecl.TokPos + 1
			importDecl.Rparen = importDecl.TokPos + 2
		}
	}

	file.Imports = append(file.Imports, newImport)
}
