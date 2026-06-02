package transformer

import (
	"go/ast"
	"strings"
	"testing"

	"github.com/gomphermp/gomphermp/internal/parser"
)

// TestTransformCritical_WrongNodeType verifies that handing transformCritical
// a directive whose Node is not a *ast.BlockStmt produces a descriptive
// error rather than panicking. The parser is supposed to populate Node with
// a BlockStmt; this test exercises the defensive branch that protects the
// transformer from a misbehaving parser.
func TestTransformCritical_WrongNodeType(t *testing.T) {
	parsed, err := parser.Parse("package main\n")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	bogus := parser.CriticalDirective{
		Node: &ast.ExprStmt{},
	}

	err = transformCritical(parsed, bogus)
	if err == nil {
		t.Fatal("expected error for non-BlockStmt Node")
	}
	if !strings.Contains(err.Error(), "expected *ast.BlockStmt") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestTransformCritical_BodyNotInAST verifies that when the directive's
// BlockStmt is not actually reachable from the file's AST (which should not
// happen with a healthy parser), the transformer reports the inconsistency
// instead of silently producing a malformed file with the runtime call
// dangling at top level.
func TestTransformCritical_BodyNotInAST(t *testing.T) {
	parsed, err := parser.Parse("package main\n")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	detached := &ast.BlockStmt{}
	bogus := parser.CriticalDirective{
		Node: detached,
	}

	err = transformCritical(parsed, bogus)
	if err == nil {
		t.Fatal("expected error when body block is detached from AST")
	}
	if !strings.Contains(err.Error(), "body block not found") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestTransform_PropagatesCriticalError verifies that when transformCritical
// fails, Transform stops and propagates the error to the caller rather than
// returning a partially transformed file. main relies on this contract to
// reject malformed inputs cleanly.
func TestTransform_PropagatesCriticalError(t *testing.T) {
	parsed, err := parser.Parse("package main\n")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	parsed.Nodes = append(parsed.Nodes, parser.AnnotatedNode{
		Directive: parser.CriticalDirective{
			Node: &ast.BlockStmt{},
		},
	})

	got, err := Transform(parsed)
	if err == nil {
		t.Fatal("expected Transform to propagate the underlying error")
	}
	if got != nil {
		t.Errorf("expected nil ParseResult on error, got %v", got)
	}
}
