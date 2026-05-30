package transformer

import (
	"go/format"
	"strings"
	"testing"

	"github.com/gomphermp/gomphermp/internal/parser"
)

// TestTransform_NoDirectives_Passthrough verifies the foundational contract:
// a file containing no //gompher directives is returned unchanged. This is
// the regression guard that protects every future directive implementation
// from accidentally modifying unrelated user code.
func TestTransform_NoDirectives_Passthrough(t *testing.T) {
	src := `package main

import "fmt"

func main() {
	fmt.Println("hello")
}
`
	parsed, err := parser.Parse(src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	transformed, err := Transform(parsed)
	if err != nil {
		t.Fatalf("transform: %v", err)
	}
	if transformed == nil {
		t.Fatal("expected non-nil ParseResult")
	}

	var buf strings.Builder
	if err := format.Node(&buf, transformed.FileSet, transformed.File); err != nil {
		t.Fatalf("format: %v", err)
	}

	if got := buf.String(); got != src {
		t.Errorf("Transform modified passthrough source.\nwant:\n%s\ngot:\n%s", src, got)
	}
}

// TestTransform_NilInput verifies that Transform handles nil input without
// panicking. The compiler's main may pass nil down degraded paths (for
// example, when an earlier stage returned an error), so the function should
// short-circuit cleanly instead of dereferencing the pointer.
func TestTransform_NilInput(t *testing.T) {
	got, err := Transform(nil)
	if err != nil {
		t.Errorf("expected nil error for nil input, got %v", err)
	}
	if got != nil {
		t.Errorf("expected nil result for nil input, got %v", got)
	}
}
