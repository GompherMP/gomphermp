package transformer

import (
	"go/ast"
	"go/token"
	"strconv"
)

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
