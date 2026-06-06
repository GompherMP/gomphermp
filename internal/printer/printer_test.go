package printer

import (
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"

	gparser "github.com/gomphermp/gomphermp/internal/parser"
	"github.com/gomphermp/gomphermp/internal/transformer"
)

// TestPrint_RoundTrip verifies that printing a parsed file produces valid Go
// source that contains the same top-level identifiers as the original.
func TestPrint_RoundTrip(t *testing.T) {
	src := `package main

import "fmt"

func main() {
	fmt.Println("hello")
}
`
	parsed, err := gparser.Parse(src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	f, err := os.CreateTemp("", "gompher_test_*.go")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	path := f.Name()
	f.Close()
	defer os.Remove(path)

	if err := Print(parsed, path); err != nil {
		t.Fatalf("Print: %v", err)
	}

	out, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	got := string(out)

	// Output must be parseable Go
	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, "", got, 0); err != nil {
		t.Errorf("output is not valid Go: %v\ngot:\n%s", err, got)
	}

	for _, want := range []string{"package main", `"fmt"`, "fmt.Println"} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in output, got:\n%s", want, got)
		}
	}
}

// TestPrint_TransformedOutput verifies that after parsing and transforming a
// file with a //gompher task directive, Print emits the runtime.Task call and
// the runtime import.
func TestPrint_TransformedOutput(t *testing.T) {
	src := `package main

func main() {
	//gompher task
	{
		work()
	}
}

func work() {}
`
	parsed, err := gparser.Parse(src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	transformed, err := transformer.Transform(parsed)
	if err != nil {
		t.Fatalf("transform: %v", err)
	}

	f, err := os.CreateTemp("", "gompher_test_*.go")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	path := f.Name()
	f.Close()
	defer os.Remove(path)

	if err := Print(transformed, path); err != nil {
		t.Fatalf("Print: %v", err)
	}

	out, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	got := string(out)

	for _, want := range []string{
		`"github.com/gomphermp/gomphermp/pkg/runtime"`,
		"runtime.Task(func() {",
		"work()",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in output, got:\n%s", want, got)
		}
	}

	// Directive comment must be gone
	if strings.Contains(got, "//gompher") {
		t.Errorf("expected directive comment removed, got:\n%s", got)
	}
}

// TestPrint_BadPath verifies that Print returns a non-nil error when the
// destination path is not writable (e.g. directory does not exist).
func TestPrint_BadPath(t *testing.T) {
	src := `package main

func main() {}
`
	parsed, err := gparser.Parse(src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	err = Print(parsed, "/nonexistent_dir_gompher/output.go")
	if err == nil {
		t.Fatal("expected error for non-writable path, got nil")
	}
}
