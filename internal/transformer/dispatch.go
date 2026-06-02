package transformer

import (
	"go/ast"
	"go/token"
)

// replaceBlockStmt walks file's AST looking for target (matched by pointer
// identity) and substitutes it with replacement. Returns true if the target
// was found and replaced.
//
// Every directive whose Node is an *ast.BlockStmt lives as one element in
// some parent BlockStmt's List slice — the only places Go puts statements.
// We walk every BlockStmt and check each of its elements against target;
// when we find the match we swap the slice entry in place and stop.
//
// Using pointer identity (==) rather than positional matching is robust: the
// parser hands us the exact pointer the directive is anchored to, so we
// cannot accidentally rewrite an unrelated block that happens to share a
// source position.
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
// dirPos from file.Comments. Comment groups remain in the file's comment
// list independently of the AST nodes they originally annotated; when a
// directive's target is replaced, its comment becomes an orphan that
// go/format renders in an awkward position (typically between the receiver
// and the method name of the synthesized call). Removing it produces clean
// output and signals that the directive has been consumed.
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
