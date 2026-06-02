package transformer

import (
	"go/ast"
	"go/token"
	"testing"

	"github.com/gomphermp/gomphermp/internal/parser"
)

// TestReplaceBlockStmt_NotFound verifies that replaceBlockStmt returns false
// when the target block is not present in the AST. This guards against
// silent failures: if the parser ever hands us a Node pointer that does not
// correspond to a real statement, transformCritical can detect it and
// surface a clear error rather than appearing to succeed.
func TestReplaceBlockStmt_NotFound(t *testing.T) {
	src := `package main

func main() {
	x := 1
	_ = x
}
`
	parsed, err := parser.Parse(src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	orphan := &ast.BlockStmt{}
	replacement := &ast.ExprStmt{X: &ast.Ident{Name: "_unused"}}

	if replaceBlockStmt(parsed.File, orphan, replacement) {
		t.Error("expected replaceBlockStmt to return false for orphan target")
	}
}

// TestRemoveDirectiveComment_SkipsDegenerateEntries verifies the defensive
// guards inside the helper: nil pointers and empty groups in File.Comments
// must be ignored rather than causing a nil dereference. The parser does
// not normally produce such entries, but future passes might temporarily
// leave them while editing the list, so the guard is real.
func TestRemoveDirectiveComment_SkipsDegenerateEntries(t *testing.T) {
	file := &ast.File{
		Comments: []*ast.CommentGroup{nil, {}, nil},
	}
	removeDirectiveComment(file, token.NoPos)
	if got := len(file.Comments); got != 3 {
		t.Errorf("expected 3 entries preserved, got %d", got)
	}
}

// TestRemoveDirectiveComment_NoMatchIsNoOp verifies that asking to remove a
// non-existent comment leaves the comment list intact. Directive handlers
// invoke this defensively after a successful node replacement, so it must
// tolerate the case where there is no comment to remove (for example, if a
// later directive type stores positions differently).
func TestRemoveDirectiveComment_NoMatchIsNoOp(t *testing.T) {
	src := `package main

// regular comment

func main() {}
`
	parsed, err := parser.Parse(src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	before := len(parsed.File.Comments)
	removeDirectiveComment(parsed.File, token.NoPos)

	if got := len(parsed.File.Comments); got != before {
		t.Errorf("expected %d comments preserved, got %d", before, got)
	}
}
