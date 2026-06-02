package transformer

import (
	"go/ast"
	"go/token"
	"strconv"
)

// runtimePkg is the local identifier used to reference the gomphermp runtime
// in transformed code. It must match the package name produced when the file
// imports github.com/gomphermp/gomphermp/pkg/runtime — Go takes the path's
// final segment as the package identifier unless an alias is declared.
const runtimePkg = "runtime"

// buildClosure wraps body in `func() { body }`. Used by every directive whose
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

// buildClosureWithIntParam wraps body in `func(name int) { body }`. Used by
// directives whose runtime callback receives an integer (Parallel's threadID,
// For's iteration index).
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

// buildRuntimeCall emits `runtime.FuncName(args...)` wrapped in an ExprStmt
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
// Used to pass directive parameters (like Critical's lock name) as string
// arguments to runtime calls.
func buildStringLit(s string) *ast.BasicLit {
	return &ast.BasicLit{
		Kind:  token.STRING,
		Value: strconv.Quote(s),
	}
}
