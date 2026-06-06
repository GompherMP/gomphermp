package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gomphermp/gomphermp/internal/parser"
	"github.com/gomphermp/gomphermp/internal/printer"
	"github.com/gomphermp/gomphermp/internal/transformer"
)

const version = "0.1.0"

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--version" {
		fmt.Println("gompher version", version)
		return
	}

	if len(os.Args) < 2 || os.Args[1] != "build" {
		fmt.Fprintln(os.Stderr, "usage: gompher build [options] <file.go>")
		fmt.Fprintln(os.Stderr, "       gompher --version")
		os.Exit(1)
	}

	runBuild(os.Args[2:])
}

func runBuild(args []string) {
	fs := flag.NewFlagSet("build", flag.ExitOnError)

	var outputFlag string
	fs.StringVar(&outputFlag, "output", "", "output binary `path`")
	fs.StringVar(&outputFlag, "o", "", "output binary `path` (shorthand)")

	var verbose bool
	fs.BoolVar(&verbose, "verbose", false, "print pipeline phases and detected directives")
	fs.BoolVar(&verbose, "v", false, "print pipeline phases and detected directives (shorthand)")

	var keepTemp bool
	fs.BoolVar(&keepTemp, "keep-temp", false, "preserve intermediate .go file after compilation")
	fs.BoolVar(&keepTemp, "k", false, "preserve intermediate .go file after compilation (shorthand)")

	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "GompherMP CLI - Structured parallelism transpiler for Go\nUsage: gompher build [options] <file.go>\n\nOptions:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if fs.NArg() != 1 {
		fs.Usage()
		os.Exit(1)
	}

	inputPath := fs.Arg(0)

	if !strings.HasSuffix(inputPath, ".go") {
		fmt.Fprintln(os.Stderr, "[Error] Input file must have a .go extension.")
		os.Exit(1)
	}

	absInput, err := filepath.Abs(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[Error] Cannot resolve path %s: %v\n", inputPath, err)
		os.Exit(1)
	}

	if _, err := os.Stat(absInput); err != nil {
		fmt.Fprintf(os.Stderr, "[Error] Cannot access %s: %v\n", inputPath, err)
		os.Exit(1)
	}

	// --- Phase 1: Read source ---

	src, err := os.ReadFile(absInput)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[Error] Cannot read %s: %v\n", inputPath, err)
		os.Exit(1)
	}

	// --- Phase 2: Parse ---

	if verbose {
		fmt.Printf("[INFO] Parsing AST of %s...\n", filepath.Base(inputPath))
	}

	parsed, err := parser.Parse(string(src))
	if err != nil {
		fmt.Fprintf(os.Stderr, "[Error] Parse failed in %s: %v\n", inputPath, err)
		os.Exit(1)
	}

	if verbose {
		for _, node := range parsed.Nodes {
			kind, line := directiveInfo(node.Directive)
			fmt.Printf("[INFO] Directive '%s' detected at line %d.\n", kind, line)
		}
	}

	// --- Phase 3: Transform ---

	transformed, err := transformer.Transform(parsed)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[Error] Transform failed in %s: %v\n", inputPath, err)
		os.Exit(1)
	}

	// --- Phase 4: Print to temp file ---

	tf, err := os.CreateTemp(filepath.Dir(absInput), "gompher_*.go")
	if err != nil {
		fmt.Fprintf(os.Stderr, "[Error] Cannot create temp file: %v\n", err)
		os.Exit(1)
	}
	tempPath := tf.Name()
	tf.Close()

	if !keepTemp {
		defer os.Remove(tempPath)
	}

	if verbose {
		fmt.Println("[INFO] Generating temporary source file...")
	}

	if err := printer.Print(transformed, tempPath); err != nil {
		fmt.Fprintf(os.Stderr, "[Error] Cannot write temp file: %v\n", err)
		os.Exit(1)
	}

	// --- Phase 5: Compile ---

	outBin := outputFlag
	if outBin == "" {
		base := strings.TrimSuffix(filepath.Base(inputPath), ".go")
		outBin = "./" + base
	}

	if verbose {
		fmt.Println("[INFO] Running go build...")
	}

	absOut, err := filepath.Abs(outBin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[Error] Cannot resolve output path: %v\n", err)
		os.Exit(1)
	}

	cmd := exec.Command("go", "build", "-o", absOut, tempPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "[Error] Compilation failed: %v\n", err)
		os.Exit(1)
	}

	if verbose {
		fmt.Printf("[SUCCESS] Binary generated: %s\n", outBin)
	}
}

// directiveInfo extracts the human-readable directive name and source line from
// any Directive value. It uses a type switch because the Directive interface
// only exposes unexported methods; the promoted Line field is accessible on
// each concrete type.
func directiveInfo(d parser.Directive) (kind string, line int) {
	switch v := d.(type) {
	case parser.ParallelDirective:
		return string(parser.DirParallel), v.Line
	case parser.ForDirective:
		return string(parser.DirFor), v.Line
	case parser.ParallelForDirective:
		return string(parser.DirParallelFor), v.Line
	case parser.SectionsDirective:
		return string(parser.DirSections), v.Line
	case parser.SectionDirective:
		return string(parser.DirSection), v.Line
	case parser.SingleDirective:
		return string(parser.DirSingle), v.Line
	case parser.MasterDirective:
		return string(parser.DirMaster), v.Line
	case parser.CriticalDirective:
		return string(parser.DirCritical), v.Line
	case parser.BarrierDirective:
		return string(parser.DirBarrier), v.Line
	case parser.AtomicDirective:
		return string(parser.DirAtomic), v.Line
	case parser.TaskDirective:
		return string(parser.DirTask), v.Line
	case parser.TaskwaitDirective:
		return string(parser.DirTaskwait), v.Line
	case parser.TaskgroupDirective:
		return string(parser.DirTaskgroup), v.Line
	case parser.TaskloopDirective:
		return string(parser.DirTaskloop), v.Line
	}
	return "unknown", 0
}
