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
// It reads raw Go source code, builds the native Abstract Syntax Tree,
// and extracts all valid //gompher directives anchored to their executable blocks.
func Parse(src string) (*ParseResult, error) {
	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("Go syntax error: %w", err)
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

// getDirectiveLine extracts the source line number from any directive.
// Used primarily to preserve the top-to-bottom execution order.
func getDirectiveLine(dir Directive) int {
	return dir.line()
}

// extractAnnotatedNodes maps //gompher directives to their corresponding Go AST nodes.
// It leverages go/ast.CommentMap to natively and accurately bind comments to executable
// statements, replacing the need for manual line-lookback heuristics.
func extractAnnotatedNodes(fset *token.FileSet, file *ast.File) ([]AnnotatedNode, error) {
	// CommentMap associates each ast.Node with the comments that physically precede it.
	cmap := ast.NewCommentMap(fset, file, file.Comments)

	var result []AnnotatedNode
	var firstErr error

	// Walk the AST looking for nodes that have GompherMP comments mapped to them.
	ast.Inspect(file, func(n ast.Node) bool {
		if n == nil || firstErr != nil {
			return false
		}

		for _, cg := range cmap[n] {
			for _, c := range cg.List {
				if !isGompherComment(c.Text) {
					continue
				}

				directive, err := parseGompherComment(fset, c)
				if err != nil {
					firstErr = err
					return false
				}

				result = append(result, AnnotatedNode{
					Directive: setNode(directive, n),
				})
			}
		}
		return true
	})

	if firstErr != nil {
		return nil, firstErr
	}

	// Preserve source file ordering to ensure the transformer processes the AST chronologically.
	sort.Slice(result, func(i, j int) bool {
		return getDirectiveLine(result[i].Directive) < getDirectiveLine(result[j].Directive)
	})

	return result, nil
}

// isGompherComment validates if a raw string is intended for the GompherMP compiler.
func isGompherComment(text string) bool {
	return strings.HasPrefix(text, "//gompher ") || text == "//gompher"
}

// parseGompherComment isolates spatial tracking from the lexical parsing of a directive.
func parseGompherComment(fset *token.FileSet, c *ast.Comment) (Directive, error) {
	line := fset.Position(c.Pos()).Line
	text := strings.TrimSpace(strings.TrimPrefix(c.Text, "//gompher"))

	directive, err := parseDirectiveText(text, c.Pos(), line)
	if err != nil {
		return nil, fmt.Errorf("line %d: %w", line, err)
	}
	return directive, nil
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
			if !strings.HasPrefix(rest, "(") || !strings.HasSuffix(rest, ")") {
				return nil, fmt.Errorf("critical name must use parentheses: critical(name)")
			}
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

	panic(fmt.Sprintf("unreachable: extractKind returned unknown kind %q", kind))
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
// Contextless or synchronization directives accept no clauses and are omitted here.
var validClauses = map[DirectiveKind][]ClauseKind{
	DirParallel:    {ClausePrivate, ClauseFirstPrivate, ClauseShared},
	DirFor:         {ClausePrivate, ClauseFirstPrivate, ClauseSchedule},
	DirParallelFor: {ClausePrivate, ClauseFirstPrivate, ClauseLastPrivate, ClauseShared, ClauseReduction, ClauseSchedule},
	DirSections:    {ClausePrivate, ClauseFirstPrivate, ClauseLastPrivate, ClauseReduction},
	DirSingle:      {ClausePrivate, ClauseFirstPrivate},
	DirTask:        {ClausePrivate, ClauseFirstPrivate, ClauseDepend},
	DirTaskloop:    {ClausePrivate, ClauseFirstPrivate, ClauseGrainsize},
}

// validateClauses cross-references extracted clauses against the validClauses map,
// enforcing structural rules before the AST mutation phase begins.
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
