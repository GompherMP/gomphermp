package transformer

import (
	"go/ast"
	"go/token"
	"strconv"
)

// This file holds the pure AST constructors: small helpers that build syntax
// nodes for the transformed program.

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

// buildRuntimeCallExpr builds the expression "runtime.FuncName(args...)" as an
// *ast.CallExpr. Use this when the call is needed as a value. Use
// buildRuntimeCall when the call stands alone as a statement.
func buildRuntimeCallExpr(funcName string, args ...ast.Expr) *ast.CallExpr {
	return &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   &ast.Ident{Name: runtimePkg},
			Sel: &ast.Ident{Name: funcName},
		},
		Args: args,
	}
}

// buildRuntimeCall emits "runtime.FuncName(args...)" wrapped in an ExprStmt
// so it can be used as a drop-in replacement for the block statement of any
// directive whose body lives inside a parent BlockStmt.
func buildRuntimeCall(funcName string, args ...ast.Expr) *ast.ExprStmt {
	return &ast.ExprStmt{X: buildRuntimeCallExpr(funcName, args...)}
}

// buildAddrOf builds the address-of expression "&e". The atomic helpers take a
// pointer to the target variable, so every atomic rewrite wraps its operand
// with this.
func buildAddrOf(e ast.Expr) *ast.UnaryExpr {
	return &ast.UnaryExpr{Op: token.AND, X: e}
}

// cloneExpr returns a deep copy of an expression with fresh (position-less)
// nodes. It is used when the same source expression must appear twice in the
// output (e.g. a loop's upper bound reused in both the worksharing call and a
// synthesized comparison): sharing a node - and its original source position -
// confuses go/format into laying the second occurrence out awkwardly.
//
// It handles the expression shapes that arise as loop bounds and variable types
// (identifiers, literals, len(x), pkg.Const, a[i], arithmetic, parenthesised
// forms, slice/array/map/chan types); for anything else it returns the node
// unchanged, which is safe because the only cost of sharing is cosmetic.
func cloneExpr(e ast.Expr) ast.Expr {
	switch v := e.(type) {
	case *ast.Ident:
		return &ast.Ident{Name: v.Name}
	case *ast.BasicLit:
		return &ast.BasicLit{Kind: v.Kind, Value: v.Value}
	case *ast.SelectorExpr:
		return &ast.SelectorExpr{X: cloneExpr(v.X), Sel: &ast.Ident{Name: v.Sel.Name}}
	case *ast.IndexExpr:
		return &ast.IndexExpr{X: cloneExpr(v.X), Index: cloneExpr(v.Index)}
	case *ast.ParenExpr:
		return &ast.ParenExpr{X: cloneExpr(v.X)}
	case *ast.StarExpr:
		return &ast.StarExpr{X: cloneExpr(v.X)}
	case *ast.UnaryExpr:
		return &ast.UnaryExpr{Op: v.Op, X: cloneExpr(v.X)}
	case *ast.BinaryExpr:
		return &ast.BinaryExpr{X: cloneExpr(v.X), Op: v.Op, Y: cloneExpr(v.Y)}
	case *ast.CallExpr:
		args := make([]ast.Expr, len(v.Args))
		for i, a := range v.Args {
			args[i] = cloneExpr(a)
		}
		return &ast.CallExpr{Fun: cloneExpr(v.Fun), Args: args}
	case *ast.ArrayType:
		return &ast.ArrayType{Len: cloneExprOrNil(v.Len), Elt: cloneExpr(v.Elt)}
	case *ast.MapType:
		return &ast.MapType{Key: cloneExpr(v.Key), Value: cloneExpr(v.Value)}
	case *ast.ChanType:
		return &ast.ChanType{Dir: v.Dir, Value: cloneExpr(v.Value)}
	case *ast.Ellipsis:
		return &ast.Ellipsis{Elt: cloneExprOrNil(v.Elt)}
	default:
		return e
	}
}

// cloneExprOrNil clones e, passing through a nil operand (e.g. an array type's
// absent length for a slice []T).
func cloneExprOrNil(e ast.Expr) ast.Expr {
	if e == nil {
		return nil
	}
	return cloneExpr(e)
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

// buildFirstprivateCapture builds "_fp_x, _fp_y := x, y" - a DEFINE
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

// buildFirstprivateShadow builds "x, y := _fp_x, _fp_y" - prepended at the
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

// binExpr, paren, intLit and atomize are small expression constructors shared by
// the codegen (notably the loop-normalization arithmetic in loopform.go).

// binExpr builds the binary expression "x op y".
func binExpr(x ast.Expr, op token.Token, y ast.Expr) *ast.BinaryExpr {
	return &ast.BinaryExpr{X: x, Op: op, Y: y}
}

// paren wraps e in parentheses: "(e)".
func paren(e ast.Expr) *ast.ParenExpr { return &ast.ParenExpr{X: e} }

// intLit builds an integer literal node with the given value.
func intLit(v string) *ast.BasicLit { return &ast.BasicLit{Kind: token.INT, Value: v} }

// atomize wraps e in parentheses when it would otherwise re-associate in a
// surrounding `*` or `/` context (binary, unary and pointer-deref expressions);
// atomic operands (identifiers, literals, calls, selectors, indexing) are left
// bare so the common case stays clean.
func atomize(e ast.Expr) ast.Expr {
	switch e.(type) {
	case *ast.BinaryExpr, *ast.UnaryExpr, *ast.StarExpr:
		return paren(e)
	default:
		return e
	}
}
