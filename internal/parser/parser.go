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

				if directiveRequiresNode(directive) {
					commentLine := fset.Position(c.Pos()).Line
					nodeLine := fset.Position(n.Pos()).Line
					if nodeLine-commentLine != 1 {
						firstErr = fmt.Errorf("line %d: directive %q must be immediately before its target (gap: %d lines)", commentLine, directive.directiveKind(), nodeLine-commentLine)
						return false
					}
					if err := validateNodeType(directive, n); err != nil {
						firstErr = fmt.Errorf("line %d: %w", commentLine, err)
						return false
					}
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

	if err := validateSectionContext(result); err != nil {
		return nil, err
	}

	// Preserve source file ordering to ensure the transformer processes the AST chronologically.
	sort.Slice(result, func(i, j int) bool {
		return getDirectiveLine(result[i].Directive) < getDirectiveLine(result[j].Directive)
	})

	return result, nil
}

// directiveRequiresNode returns false for directives that are pure synchronization points
// (barrier, taskwait) and have no associated executable block.
func directiveRequiresNode(dir Directive) bool {
	switch dir.directiveKind() {
	case DirBarrier, DirTaskwait:
		return false
	}
	return true
}

// validateNodeType enforces that each directive is attached to the correct Go AST node kind.
// This catches user errors like //gompher for placed over a non-loop statement before the
// transformer attempts a type assertion that would otherwise panic at runtime.
func validateNodeType(dir Directive, node ast.Node) error {
	switch dir.directiveKind() {
	case DirFor, DirParallelFor, DirTaskloop:
		if _, ok := node.(*ast.ForStmt); !ok {
			return fmt.Errorf("directive %q requires a for loop, got %T", dir.directiveKind(), node)
		}
	case DirAtomic:
		switch node.(type) {
		case *ast.ExprStmt, *ast.AssignStmt, *ast.IncDecStmt:
			// valid
		default:
			return fmt.Errorf("directive %q requires an expression or assignment statement, got %T", dir.directiveKind(), node)
		}
	case DirParallel, DirSections, DirSection, DirSingle, DirMaster, DirCritical, DirTask, DirTaskgroup:
		if _, ok := node.(*ast.BlockStmt); !ok {
			return fmt.Errorf("directive %q requires a block statement, got %T", dir.directiveKind(), node)
		}
	}
	return nil
}

// validateSectionContext enforces that every //gompher section directive lives inside a
// //gompher sections directive's block. It uses source position containment to check membership.
func validateSectionContext(result []AnnotatedNode) error {
	var sectionsNodes []ast.Node
	for _, n := range result {
		if d, ok := n.Directive.(SectionsDirective); ok {
			sectionsNodes = append(sectionsNodes, d.Node)
		}
	}

	for _, n := range result {
		d, ok := n.Directive.(SectionDirective)
		if !ok {
			continue
		}

		inside := false
		for _, s := range sectionsNodes {
			if d.Node.Pos() >= s.Pos() && d.Node.End() <= s.End() {
				inside = true
				break
			}
		}
		if !inside {
			return fmt.Errorf("line %d: directive %q must appear inside a //gompher sections block", d.Line, DirSection)
		}
	}
	return nil
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
// It acts as the orchestrator: identifying the kind, then delegating construction.
func parseDirectiveText(text string, p token.Pos, line int) (Directive, error) {
	if text == "" {
		return nil, fmt.Errorf("empty //gompher directive")
	}

	kind, rest, err := extractKind(text)
	if err != nil {
		return nil, err
	}

	return buildDirective(kind, rest, pos{Pos: p, Line: line})
}

// buildDirective constructs the concrete Directive for an already-validated kind.
// Separated from parseDirectiveText so the terminal error return is reachable from tests.
func buildDirective(kind DirectiveKind, rest string, srcPos pos) (Directive, error) {
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
// Contextless or synchronization directives accept no clauses and are omitted here.
var validClauses = map[DirectiveKind][]ClauseKind{
	DirParallel:    {ClausePrivate, ClauseFirstPrivate, ClauseShared},
	DirFor:         {ClausePrivate, ClauseFirstPrivate, ClauseSchedule},
	DirParallelFor: {ClausePrivate, ClauseFirstPrivate, ClauseLastPrivate, ClauseShared, ClauseReduction, ClauseSchedule},
	DirSections:    {ClausePrivate, ClauseFirstPrivate, ClauseLastPrivate, ClauseReduction},
	DirSingle:      {ClausePrivate, ClauseFirstPrivate},
	DirTask:        {ClausePrivate, ClauseFirstPrivate, ClauseShared, ClauseReduction, ClauseDepend},
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
