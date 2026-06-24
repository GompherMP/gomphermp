package transformer

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/gomphermp/gomphermp/internal/parser"
)

// This file holds the AST mutators: helpers that walk the file's syntax tree to
// substitute a directive's target node with the synthesized runtime call, insert
// statements at a position, or strip a consumed directive comment.

// replaceBlockStmt walks file's AST looking for target and substitutes it with
// replacement. Returns true if the target was found and replaced. Every
// directive whose Node is an *ast.BlockStmt lives as one element in some parent
// BlockStmt's List slice; we walk every BlockStmt and check each element against
// target.
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

// replaceStmt swaps target for replacement in whatever parent BlockStmt holds
// it, matched by pointer identity. It generalizes replaceBlockStmt and
// replaceForStmt to any statement type, which the atomic rewrite needs because
// its targets are *ast.IncDecStmt and *ast.AssignStmt rather than blocks.
func replaceStmt(file *ast.File, target, replacement ast.Stmt) bool {
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
			if stmt == target {
				block.List[i] = replacement
				replaced = true
				return false
			}
		}
		return true
	})
	return replaced
}

// replaceStmtWithPrefix swaps target for replacement in its parent block and
// also inserts prefix statements just before it, matched by pointer identity.
// It generalizes replaceBlockStmtWithPrefix to any statement type (the loop
// data-clause rewrite needs it to emit firstprivate/reduction captures
// immediately before the synthesized parallel call).
func replaceStmtWithPrefix(file *ast.File, target, replacement ast.Stmt, prefix []ast.Stmt) bool {
	if len(prefix) == 0 {
		return replaceStmt(file, target, replacement)
	}
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
			if stmt == target {
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

// replaceForStmt walks file's AST looking for target *ast.ForStmt in a parent
// BlockStmt's List and substitutes it with replacement. Returns true if found.
func replaceForStmt(file *ast.File, target *ast.ForStmt, replacement ast.Stmt) bool {
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
			if fs, ok := stmt.(*ast.ForStmt); ok && fs == target {
				block.List[i] = replacement
				replaced = true
				return false
			}
		}
		return true
	})
	return replaced
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

// removeDirectiveComment strips the //gompher comment group anchored at dirPos
// from file.Comments.
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

// transformBlockDirective implements the shared rewrite pattern used by every
// directive whose Node is a *ast.BlockStmt and whose runtime entry point accepts
// a parameterless closure: Critical, Master.
func transformBlockDirective(
	file *parser.ParseResult,
	node ast.Node,
	dirPos token.Pos,
	line int,
	runtimeFunc string,
	prefixArgs ...ast.Expr,
	// prefixArgs are the arguments that come before the closure in the runtime call.
	// Critical passes a string literal (the lock name), Master passes the threadID
	// identifier from the enclosing parallel scope.
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
