// Package printer converts a parsed (and optionally transformed) Go AST back
// into gofmt-canonical Go source and writes it to a file on disk.
package printer

import (
	"go/format"
	"os"

	"github.com/gomphermp/gomphermp/internal/parser"
)

// Print formats result's AST using gofmt rules and writes the output to the
// file at path, creating or truncating it as needed.
func Print(result *parser.ParseResult, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return format.Node(f, result.FileSet, result.File)
}
