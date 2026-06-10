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

// buildFuncSlice builds a `[]func(){ elems... }` composite literal. Used by the
// sections directive to pass its per-section closures to runtime.Sections,
// whose parameter type is []func().
func buildFuncSlice(elems []ast.Expr) *ast.CompositeLit {
	return &ast.CompositeLit{
		Type: &ast.ArrayType{
			Elt: &ast.FuncType{Params: &ast.FieldList{}},
		},
		Elts: elems,
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

// replaceBlockStmtWithPrefix walks file looking for the parent BlockStmt that
// contains target as a direct child, then atomically replaces target with
// prefix stmts followed by replacement. Used by firstprivate to inject the
// capture assignment (_fp_x := x) immediately before the Task call in one pass.
func replaceBlockStmtWithPrefix(file *ast.File, target *ast.BlockStmt, replacement ast.Stmt, prefix []ast.Stmt) bool {
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
				newList := make([]ast.Stmt, 0, len(block.List)+len(prefix))
				newList = append(newList, block.List[:i]...)
				newList = append(newList, prefix...)
				newList = append(newList, replacement)
				newList = append(newList, block.List[i+1:]...)
				block.List = newList
				replaced = true
				return false
			}
		}
		return true
	})
	return replaced
}

// findVarType walks file collecting all explicit-type declarations of varName
// whose position precedes beforePos, and returns the type expression of the
// closest (latest) one. Handles var declarations and function parameters.
// Returns a descriptive error when no explicit-type declaration is found,
// guiding the user to add one.
func findVarType(file *ast.File, varName string, beforePos token.Pos) (ast.Expr, error) {
	var found ast.Expr
	var foundPos token.Pos

	ast.Inspect(file, func(n ast.Node) bool {
		if n == nil {
			return false
		}
		switch decl := n.(type) {
		case *ast.GenDecl:
			if decl.Tok != token.VAR {
				return true
			}
			for _, spec := range decl.Specs {
				vs, ok := spec.(*ast.ValueSpec)
				if !ok || vs.Type == nil || vs.Pos() >= beforePos {
					continue
				}
				for _, name := range vs.Names {
					if name.Name == varName && vs.Pos() > foundPos {
						found = vs.Type
						foundPos = vs.Pos()
					}
				}
			}
		case *ast.FuncDecl:
			if decl.Type == nil || decl.Type.Params == nil {
				return true
			}
			for _, field := range decl.Type.Params.List {
				if field.Type == nil || field.Pos() >= beforePos {
					continue
				}
				for _, name := range field.Names {
					if name.Name == varName && field.Pos() > foundPos {
						found = field.Type
						foundPos = field.Pos()
					}
				}
			}
		}
		return true
	})

	if found == nil {
		return nil, fmt.Errorf("private clause: cannot determine type of %q — use an explicit 'var %s T' declaration", varName, varName)
	}
	return found, nil
}

// buildPrivateVarDecl builds "var name T" as a DeclStmt using typeExpr
// verbatim from the AST. The zero value of T shadows any outer variable of
// the same name inside the closure.
func buildPrivateVarDecl(name string, typeExpr ast.Expr) *ast.DeclStmt {
	return &ast.DeclStmt{
		Decl: &ast.GenDecl{
			Tok: token.VAR,
			Specs: []ast.Spec{
				&ast.ValueSpec{
					Names: []*ast.Ident{{Name: name}},
					Type:  typeExpr,
				},
			},
		},
	}
}

// buildFirstprivateCapture builds "_fp_x, _fp_y := x, y" — a DEFINE
// assignment that snapshots the current values of vars before the goroutine
// starts. Injected into the surrounding block immediately before the Task call.
func buildFirstprivateCapture(vars []string) *ast.AssignStmt {
	lhs := make([]ast.Expr, len(vars))
	rhs := make([]ast.Expr, len(vars))
	for i, v := range vars {
		lhs[i] = &ast.Ident{Name: "_fp_" + v}
		rhs[i] = &ast.Ident{Name: v}
	}
	return &ast.AssignStmt{Lhs: lhs, Tok: token.DEFINE, Rhs: rhs}
}

// buildFirstprivateShadow builds "x, y := _fp_x, _fp_y" — prepended at the
// top of the closure body so the original names refer to the captured copies
// rather than the outer variables captured by reference.
func buildFirstprivateShadow(vars []string) *ast.AssignStmt {
	lhs := make([]ast.Expr, len(vars))
	rhs := make([]ast.Expr, len(vars))
	for i, v := range vars {
		lhs[i] = &ast.Ident{Name: v}
		rhs[i] = &ast.Ident{Name: "_fp_" + v}
	}
	return &ast.AssignStmt{Lhs: lhs, Tok: token.DEFINE, Rhs: rhs}
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
