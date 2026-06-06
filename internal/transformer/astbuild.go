package transformer

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"

	"github.com/gomphermp/gomphermp/internal/parser"
)

const unsafeImportPath = "unsafe"

// runtimePkg is the local identifier used to reference the gomphermp runtime
// in transformed code.
const runtimePkg = "runtime"

// buildClosure wraps body in "func() { body }". Used by every directive whose
// runtime entry point accepts a parameterless callback (Critical, Single,
// Sections elements, Task variants).
func buildClosure(body *ast.BlockStmt) *ast.FuncLit {
	return &ast.FuncLit{
		Type: &ast.FuncType{
			Params: &ast.FieldList{},
		},
		Body: body,
	}
}

// buildClosureWithIntParam wraps body in "func(name int) { body }". Used by
// directives whose runtime callback receives an integer (Parallel's threadID,
// For's iteration index, etc.).
func buildClosureWithIntParam(body *ast.BlockStmt, name string) *ast.FuncLit {
	return &ast.FuncLit{
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{{Name: name}},
						Type:  &ast.Ident{Name: "int"},
					},
				},
			},
		},
		Body: body,
	}
}

// buildRuntimeCall emits "runtime.FuncName(args...)" wrapped in an ExprStmt
// so it can be used as a drop-in replacement for the block statement of any
// directive whose body lives inside a parent BlockStmt.
func buildRuntimeCall(funcName string, args ...ast.Expr) *ast.ExprStmt {
	return &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.Ident{Name: runtimePkg},
				Sel: &ast.Ident{Name: funcName},
			},
			Args: args,
		},
	}
}

// buildStringLit produces a quoted Go string literal node from a raw string.
// Used to pass directive parameters as string arguments to runtime calls.
func buildStringLit(s string) *ast.BasicLit {
	return &ast.BasicLit{
		Kind:  token.STRING,
		Value: strconv.Quote(s),
	}
}

// replaceBlockStmt walks file's AST looking for target and substitutes
// it with replacement. Returns true if the target
// was found and replaced. Every directive whose Node is an *ast.BlockStmt
// lives as one element in some parent BlockStmt's List slice. We walk every
// BlockStmt and check each of its elements against target
func replaceBlockStmt(file *ast.File, target *ast.BlockStmt, replacement ast.Stmt) bool {
	var replaced bool
	ast.Inspect(file, func(n ast.Node) bool {
		if replaced {
			return false
		}
		block, ok := n.(*ast.BlockStmt)
		if !ok {
			return true
		}
		for i, stmt := range block.List {
			if inner, ok := stmt.(*ast.BlockStmt); ok && inner == target {
				block.List[i] = replacement
				replaced = true
				return false
			}
		}
		return true
	})
	return replaced
}

// transformBlockDirective implements the shared rewrite pattern used by
// every directive whose Node is a *ast.BlockStmt and whose runtime entry
// point accepts a parameterless closure: Critical, Single, Master.

// prefixArgs are the arguments that come before the closure in the runtime
// call. Critical passes a string literal (the lock name), Master passes the
// threadID identifier from the enclosing parallel scope, Single passes
// nothing.
func transformBlockDirective(
	file *parser.ParseResult,
	node ast.Node,
	dirPos token.Pos,
	line int,
	runtimeFunc string,
	prefixArgs ...ast.Expr,
) error {
	body, ok := node.(*ast.BlockStmt)
	if !ok {
		return fmt.Errorf("%s at line %d: expected *ast.BlockStmt, got %T", runtimeFunc, line, node)
	}

	args := append(prefixArgs, buildClosure(body))
	call := buildRuntimeCall(runtimeFunc, args...)

	if !replaceBlockStmt(file.File, body, call) {
		return fmt.Errorf("%s at line %d: body block not found in AST", runtimeFunc, line)
	}
	removeDirectiveComment(file.File, dirPos)

	return nil
}

// insertCallAtPos finds the innermost *ast.BlockStmt in file that contains
// dirPos (i.e. Lbrace < dirPos < Rbrace) and inserts call at the correct
// position within that block: before the first statement that begins after
// dirPos, or appended at the end if the directive is the last thing in the
// block. Used by point directives that have no associated AST node (Taskwait,
// Barrier).
func insertCallAtPos(file *ast.File, dirPos token.Pos, call ast.Stmt) bool {
	var candidates []*ast.BlockStmt
	ast.Inspect(file, func(n ast.Node) bool {
		block, ok := n.(*ast.BlockStmt)
		if !ok {
			return true
		}
		if block.Lbrace < dirPos && dirPos < block.Rbrace {
			candidates = append(candidates, block)
		}
		return true
	})

	if len(candidates) == 0 {
		return false
	}

	innermost := candidates[0]
	for _, b := range candidates[1:] {
		if b.Lbrace > innermost.Lbrace {
			innermost = b
		}
	}

	for i, stmt := range innermost.List {
		if stmt.Pos() > dirPos {
			newList := make([]ast.Stmt, 0, len(innermost.List)+1)
			newList = append(newList, innermost.List[:i]...)
			newList = append(newList, call)
			newList = append(newList, innermost.List[i:]...)
			innermost.List = newList
			return true
		}
	}
	innermost.List = append(innermost.List, call)
	return true
}

// buildUintptrSlice builds the AST for []uintptr{uintptr(unsafe.Pointer(&v))...}.
// Returns a nil identifier when vars is empty so the call site stays readable.
func buildUintptrSlice(vars []string) ast.Expr {
	if len(vars) == 0 {
		return &ast.Ident{Name: "nil"}
	}
	elts := make([]ast.Expr, len(vars))
	for i, v := range vars {
		elts[i] = buildUintptrOfVar(v)
	}
	return &ast.CompositeLit{
		Type: &ast.ArrayType{Elt: &ast.Ident{Name: "uintptr"}},
		Elts: elts,
	}
}

// buildUintptrOfVar returns the AST for uintptr(unsafe.Pointer(&name)).
func buildUintptrOfVar(name string) ast.Expr {
	return &ast.CallExpr{
		Fun: &ast.Ident{Name: "uintptr"},
		Args: []ast.Expr{
			&ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "unsafe"},
					Sel: &ast.Ident{Name: "Pointer"},
				},
				Args: []ast.Expr{
					&ast.UnaryExpr{
						Op: token.AND,
						X:  &ast.Ident{Name: name},
					},
				},
			},
		},
	}
}

// ensureUnsafeImport adds an import of "unsafe" to file if not already present.
// Follows the same idempotent pattern as ensureRuntimeImport.
func ensureUnsafeImport(file *ast.File) {
	quoted := strconv.Quote(unsafeImportPath)

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

// removeDirectiveComment strips the //gompher comment group anchored at
// dirPos from file.Comments.
func removeDirectiveComment(file *ast.File, dirPos token.Pos) {
	for i, cg := range file.Comments {
		if cg == nil || len(cg.List) == 0 {
			continue
		}
		if cg.List[0].Slash == dirPos {
			file.Comments = append(file.Comments[:i], file.Comments[i+1:]...)
			return
		}
	}
}
