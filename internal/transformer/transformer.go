// Package transformer rewrites a parsed Go program that contains //gompher
// directives into an equivalent program that calls the gomphermp runtime
// instead.
package transformer

import (
	"go/ast"
	"go/token"
	"strconv"
	"github.com/gomphermp/gomphermp/internal/parser"
)

const runtimeImportPath = "github.com/gomphermp/gomphermp/pkg/runtime"

// Transform rewrites every annotated GompherMP directive in the parsed file
// into the corresponding runtime call. Nodes that do not correspond to a
// directive are passed through unmodified.
func Transform(result *parser.ParseResult) (*parser.ParseResult, error) {
	if result == nil {
		return nil, nil
	}

	var emittedRuntime bool
	for _, node := range result.Nodes {
		switch d := node.Directive.(type) {
		case parser.ParallelDirective:
			if err := transformParallel(result, d); err != nil {
				return nil, err
			}
			emittedRuntime = true
		case parser.ForDirective:
			if err := transformFor(result, d); err != nil {
				return nil, err
			}
			emittedRuntime = true
		case parser.ParallelForDirective:
			if err := transformParallelFor(result, d); err != nil {
				return nil, err
			}
			emittedRuntime = true
		case parser.SectionsDirective:
			if err := transformSections(result, d); err != nil {
				return nil, err
			}
			emittedRuntime = true
		case parser.ParallelSectionsDirective:
			if err := transformParallelSections(result, d); err != nil {
				return nil, err
			}
			emittedRuntime = true
		case parser.BarrierDirective:
			if err := transformBarrier(result, d); err != nil {
				return nil, err
			}
			emittedRuntime = true
		case parser.AtomicDirective:
			if err := transformAtomic(result, d); err != nil {
				return nil, err
			}
			emittedRuntime = true
		case parser.CriticalDirective:
			if err := transformCritical(result, d); err != nil {
				return nil, err
			}
			emittedRuntime = true
		case parser.SingleDirective:
			if err := transformSingle(result, d); err != nil {
				return nil, err
			}
			emittedRuntime = true
		case parser.MasterDirective:
			if err := transformMaster(result, d); err != nil {
				return nil, err
			}
			emittedRuntime = true
		case parser.TaskDirective:
			if err := transformTask(result, d); err != nil {
				return nil, err
			}
			emittedRuntime = true
		case parser.TaskwaitDirective:
			if err := transformTaskwait(result, d); err != nil {
				return nil, err
			}
			emittedRuntime = true
		case parser.TaskgroupDirective:
			if err := transformTaskgroup(result, d); err != nil {
				return nil, err
			}
			emittedRuntime = true
		case parser.TaskloopDirective:
			if err := transformTaskloop(result, d); err != nil {
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

// ensureRuntimeImport adds an import of the gomphermp runtime to file if it
// is not already imported. The operation is idempotent: invoking it multiple
// times on the same file has no additional effect, which lets directive
// handlers call it unconditionally without coordinating among themselves.
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
		if importDecl.Lparen == token.NoPos {
			importDecl.Lparen = importDecl.TokPos + 1
			importDecl.Rparen = importDecl.TokPos + 2
		}
	}

	file.Imports = append(file.Imports, newImport)
}
