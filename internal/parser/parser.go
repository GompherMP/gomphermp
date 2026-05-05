package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"sort"
	"strings"
)

// AnnotatedNode pairs a parsed GompherMP directive with its corresponding Go syntax node.
type AnnotatedNode struct {
	Directive Directive // concrete type: ParallelDirective, ForDirective, etc.
}

// ParseResult holds the complete context of a parsed Go file,
// including the spatial data (FileSet) and the extracted GompherMP nodes.
type ParseResult struct {
	FileSet *token.FileSet
	File    *ast.File
	Nodes   []AnnotatedNode
}

// Parse is the main entry point for the compiler frontend.
// It reads raw Go source code, builds the native Abstract Syntax Tree (AST),
// and extracts all valid //gompher directives anchored to their executable blocks.
func Parse(src string) (*ParseResult, error) {
	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("Go sintaxis error: %w", err)
	}

	nodes, err := extractAnnotatedNodes(fset, file)
	if err != nil {
		return nil, err
	}

	return &ParseResult{
		FileSet: fset,
		File:    file,
		Nodes:   nodes,
	}, nil
}

// setNode injects the target Go AST node into the concrete directive struct.
// Because Directive is an interface holding value types, a type switch is
// required to return a modified copy of the underlying struct.
func setNode(dir Directive, node ast.Node) Directive {
	switch d := dir.(type) {
	case ParallelDirective:
		d.Node = node
		return d
	case ForDirective:
		d.Node = node
		return d
	case ParallelForDirective:
		d.Node = node
		return d
	case SectionsDirective:
		d.Node = node
		return d
	case SectionDirective:
		d.Node = node
		return d
	case SingleDirective:
		d.Node = node
		return d
	case MasterDirective:
		d.Node = node
		return d
	case CriticalDirective:
		d.Node = node
		return d
	case AtomicDirective:
		d.Node = node
		return d
	case TaskDirective:
		d.Node = node
		return d
	case TaskgroupDirective:
		d.Node = node
		return d
	case TaskloopDirective:
		d.Node = node
		return d
	default:
		return dir // BarrierDirective, TaskwaitDirective — no node required
	}
}

// getDirectiveLine safely extracts the original source line number from any directive.
func getDirectiveLine(dir Directive) int {
	switch d := dir.(type) {
	case ParallelDirective:
		return d.Line
	case ForDirective:
		return d.Line
	case ParallelForDirective:
		return d.Line
	case SectionsDirective:
		return d.Line
	case SectionDirective:
		return d.Line
	case SingleDirective:
		return d.Line
	case MasterDirective:
		return d.Line
	case CriticalDirective:
		return d.Line
	case BarrierDirective:
		return d.Line
	case AtomicDirective:
		return d.Line
	case TaskDirective:
		return d.Line
	case TaskwaitDirective:
		return d.Line
	case TaskgroupDirective:
		return d.Line
	case TaskloopDirective:
		return d.Line
	default:
		return 0
	}
}

// extractAnnotatedNodes maps isolated //gompher comments to their adjacent Go AST nodes.
// It utilizes a 3-line lookback heuristic to bind comments to execution blocks.
func extractAnnotatedNodes(fset *token.FileSet, file *ast.File) ([]AnnotatedNode, error) {
	var result []AnnotatedNode

	// Map line number -> parsed directive for fast lookup during AST traversal.
	directiveByLine := make(map[int]Directive)

	// 1. Scan all comments to find and parse standalone GompherMP directives.
	for _, cg := range file.Comments {
		for _, c := range cg.List {
			if !strings.HasPrefix(c.Text, "//gompher ") && c.Text != "//gompher" {
				continue
			}

			line := fset.Position(c.Pos()).Line
			text := strings.TrimSpace(strings.TrimPrefix(c.Text, "//gompher"))

			directive, err := parseDirectiveText(text, c.Pos(), line)
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", line, err)
			}

			directiveByLine[line] = directive
		}
	}

	// Directives that act as pure synchronization points and do not wrap code blocks.
	noNodeDirectives := map[DirectiveKind]bool{
		DirBarrier:  true,
		DirTaskwait: true,
	}

	matched := make(map[int]bool)

	// 2. Traverse the Go AST and apply the lookback algorithm.
	ast.Inspect(file, func(n ast.Node) bool {
		if n == nil {
			return false
		}

		block, ok := n.(*ast.BlockStmt)
		if !ok {
			return true
		}

		// Look up to 3 lines above the current block statement to find its directive.
		for _, stmt := range block.List {
			stmtLine := fset.Position(stmt.Pos()).Line

			for lookback := 1; lookback <= 3; lookback++ {
				dir, exists := directiveByLine[stmtLine-lookback]
				if !exists {
					continue
				}
				if matched[stmtLine-lookback] {
					break
				}
				matched[stmtLine-lookback] = true
				result = append(result, AnnotatedNode{
					Directive: setNode(dir, stmt),
				})
				break
			}
		}

		return true
	})

	// 3. Append contextless directives (like barrier) that were not matched to a block.
	for line, dir := range directiveByLine {
		if noNodeDirectives[dir.directiveKind()] && !matched[line] {
			result = append(result, AnnotatedNode{
				Directive: dir,
			})
		}
	}

	// 4. Ensure top-to-bottom execution order is preserved.
	sort.Slice(result, func(i, j int) bool {
		return getDirectiveLine(result[i].Directive) < getDirectiveLine(result[j].Directive)
	})

	return result, nil
}

// parseDirectiveText translates a raw comment string into a strongly-typed Directive struct.
// It acts as the orchestrator: identifying the kind, extracting clauses, and running validation.
func parseDirectiveText(text string, p token.Pos, line int) (Directive, error) {
	if text == "" {
		return nil, fmt.Errorf("empty //gompher directive")
	}

	kind, rest, err := extractKind(text)
	if err != nil {
		return nil, err
	}

	srcPos := pos{Pos: p, Line: line}

	switch kind {
	case DirParallel:
		clauses, err := extractClauses(rest)
		if err != nil {
			return nil, err
		}
		if err := validateClauses(kind, clauses); err != nil {
			return nil, err
		}
		return ParallelDirective{Clauses: clauses, pos: srcPos}, nil

	case DirFor:
		clauses, err := extractClauses(rest)
		if err != nil {
			return nil, err
		}
		if err := validateClauses(kind, clauses); err != nil {
			return nil, err
		}
		return ForDirective{Clauses: clauses, pos: srcPos}, nil

	case DirParallelFor:
		clauses, err := extractClauses(rest)
		if err != nil {
			return nil, err
		}
		if err := validateClauses(kind, clauses); err != nil {
			return nil, err
		}
		return ParallelForDirective{Clauses: clauses, pos: srcPos}, nil

	case DirSections:
		clauses, err := extractClauses(rest)
		if err != nil {
			return nil, err
		}
		if err := validateClauses(kind, clauses); err != nil {
			return nil, err
		}
		return SectionsDirective{Clauses: clauses, pos: srcPos}, nil

	case DirSection:
		if rest != "" {
			return nil, fmt.Errorf("directive %q accepts no clauses", kind)
		}
		return SectionDirective{pos: srcPos}, nil

	case DirSingle:
		clauses, err := extractClauses(rest)
		if err != nil {
			return nil, err
		}
		if err := validateClauses(kind, clauses); err != nil {
			return nil, err
		}
		return SingleDirective{Clauses: clauses, pos: srcPos}, nil

	case DirMaster:
		if rest != "" {
			return nil, fmt.Errorf("directive %q accepts no clauses", kind)
		}
		return MasterDirective{pos: srcPos}, nil

	case DirCritical:
		name := ""
		if rest != "" {
			name = strings.Trim(rest, "()")
			if name == "" {
				return nil, fmt.Errorf("critical name cannot be empty")
			}
		}
		return CriticalDirective{Name: name, pos: srcPos}, nil

	case DirBarrier:
		if rest != "" {
			return nil, fmt.Errorf("directive %q accepts no clauses", kind)
		}
		return BarrierDirective{pos: srcPos}, nil

	case DirAtomic:
		mode := ""
		if rest != "" {
			valid := map[string]bool{"read": true, "write": true, "update": true}
			if !valid[rest] {
				return nil, fmt.Errorf("invalid atomic mode: %q", rest)
			}
			mode = rest
		}
		return AtomicDirective{Mode: mode, pos: srcPos}, nil

	case DirTask:
		clauses, err := extractClauses(rest)
		if err != nil {
			return nil, err
		}
		if err := validateClauses(kind, clauses); err != nil {
			return nil, err
		}
		return TaskDirective{Clauses: clauses, pos: srcPos}, nil

	case DirTaskwait:
		if rest != "" {
			return nil, fmt.Errorf("directive %q accepts no clauses", kind)
		}
		return TaskwaitDirective{pos: srcPos}, nil

	case DirTaskgroup:
		if rest != "" {
			return nil, fmt.Errorf("directive %q accepts no clauses", kind)
		}
		return TaskgroupDirective{pos: srcPos}, nil

	case DirTaskloop:
		clauses, err := extractClauses(rest)
		if err != nil {
			return nil, err
		}
		if err := validateClauses(kind, clauses); err != nil {
			return nil, err
		}
		return TaskloopDirective{Clauses: clauses, pos: srcPos}, nil
	}

	return nil, fmt.Errorf("unknown directive: %q", kind)
}

// extractKind matches the base directive name.
// The 'kinds' array is strictly ordered to evaluate composite names (like "parallel for")
// before partial names ("parallel" or "for") to prevent premature matching.
func extractKind(text string) (DirectiveKind, string, error) {
	kinds := []DirectiveKind{
		DirParallelFor,
		DirParallel,
		DirFor,
		DirSections,
		DirSection,
		DirSingle,
		DirMaster,
		DirCritical,
		DirBarrier,
		DirAtomic,
		DirTask,
		DirTaskwait,
		DirTaskgroup,
		DirTaskloop,
	}

	for _, kind := range kinds {
		prefix := string(kind)
		if text == prefix {
			return kind, "", nil
		}
		if strings.HasPrefix(text, prefix+" ") {
			rest := strings.TrimPrefix(text, prefix+" ")
			return kind, strings.TrimSpace(rest), nil
		}

		if strings.HasPrefix(text, prefix+"(") {
			rest := strings.TrimPrefix(text, prefix)
			return kind, strings.TrimSpace(rest), nil
		}
	}

	return "", "", fmt.Errorf("unknown directive: %q", text)
}

// validClauses defines the strict compliance mapping for OpenMP directives.
// Directives not present in this map are treated as synchronization constructs
// that do not accept standard data-sharing clauses.
var validClauses = map[DirectiveKind][]ClauseKind{
	DirParallel:    {ClausePrivate, ClauseFirstPrivate, ClauseShared},
	DirFor:         {ClausePrivate, ClauseFirstPrivate, ClauseSchedule},
	DirParallelFor: {ClausePrivate, ClauseFirstPrivate, ClauseLastPrivate, ClauseShared, ClauseReduction, ClauseSchedule},
	DirSections:    {ClausePrivate, ClauseFirstPrivate, ClauseLastPrivate, ClauseReduction},
	DirSingle:      {ClausePrivate, ClauseFirstPrivate},
	DirTask:        {ClausePrivate, ClauseFirstPrivate, ClauseDepend, ClauseReduction},
	DirTaskloop:    {ClausePrivate, ClauseFirstPrivate, ClauseGrainsize},
	// Contextless or sync directives accept no clauses:
	// DirSection, DirMaster, DirBarrier, DirAtomic, DirCritical, DirTaskwait, DirTaskgroup
}

// validateClauses cross-references extracted clauses against the validClauses map,
// enforcing OpenMP structural rules before the compilation continues.
func validateClauses(kind DirectiveKind, clauses []Clause) error {
	allowed, exists := validClauses[kind]

	if !exists {
		if len(clauses) > 0 {
			return fmt.Errorf("directive %q accepts no clauses", kind)
		}
		return nil
	}

	for _, clause := range clauses {
		ck := clause.clauseKind()
		found := false
		for _, a := range allowed {
			if ck == a {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("directive %q does not accept clause %q", kind, ck)
		}
	}

	return nil
}
