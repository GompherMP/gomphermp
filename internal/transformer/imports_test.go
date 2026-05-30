package transformer

import (
	"go/ast"
	"go/token"
	"testing"

	"github.com/gomphermp/gomphermp/internal/parser"
)

// TestEnsureRuntimeImport_AddsToFileWithoutImports verifies that the helper
// creates a brand-new import declaration when the file does not import
// anything. This is the simplest input shape and the one that exercises the
// "no existing block" branch of the helper.
func TestEnsureRuntimeImport_AddsToFileWithoutImports(t *testing.T) {
	src := `package main

func main() {}
`
	parsed, err := parser.Parse(src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	ensureRuntimeImport(parsed.File)

	if len(parsed.File.Imports) != 1 {
		t.Fatalf("expected 1 import, got %d", len(parsed.File.Imports))
	}
	if got := parsed.File.Imports[0].Path.Value; got != `"github.com/gomphermp/gomphermp/pkg/runtime"` {
		t.Errorf("wrong import path: %s", got)
	}
}

// TestEnsureRuntimeImport_Idempotent verifies that invoking the helper twice
// in a row does not produce a duplicate import. Directive handlers will call
// it unconditionally before emitting their runtime calls, so duplicate
// protection must live in the helper itself.
func TestEnsureRuntimeImport_Idempotent(t *testing.T) {
	src := `package main

import "github.com/gomphermp/gomphermp/pkg/runtime"

func main() { _ = runtime.PoolSize() }
`
	parsed, err := parser.Parse(src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	ensureRuntimeImport(parsed.File)
	ensureRuntimeImport(parsed.File)

	count := 0
	for _, imp := range parsed.File.Imports {
		if imp.Path.Value == `"github.com/gomphermp/gomphermp/pkg/runtime"` {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected 1 runtime import after double invocation, got %d", count)
	}
}

// TestEnsureRuntimeImport_AppendsToExistingBlock verifies that when the file
// already has an import declaration, the runtime import is appended to it
// instead of creating a parallel declaration. Keeping imports consolidated
// in a single block matches gofmt's canonical layout.
func TestEnsureRuntimeImport_AppendsToExistingBlock(t *testing.T) {
	src := `package main

import "fmt"

func main() { fmt.Println("hi") }
`
	parsed, err := parser.Parse(src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	ensureRuntimeImport(parsed.File)

	importDecls := 0
	for _, decl := range parsed.File.Decls {
		if gd, ok := decl.(*ast.GenDecl); ok && gd.Tok == token.IMPORT {
			importDecls++
		}
	}
	if importDecls != 1 {
		t.Errorf("expected import declarations consolidated into 1, got %d", importDecls)
	}
	if len(parsed.File.Imports) != 2 {
		t.Errorf("expected 2 imports after append, got %d", len(parsed.File.Imports))
	}
}
