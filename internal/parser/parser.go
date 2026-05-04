package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"sort"
	"strings"
)

type AnnotatedNode struct {
	Directive *GompherDirective
	Node      ast.Node
}

type ParseResult struct {
	FileSet *token.FileSet
	File    *ast.File
	Nodes   []AnnotatedNode
}

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

func extractAnnotatedNodes(fset *token.FileSet, file *ast.File) ([]AnnotatedNode, error) {
	var result []AnnotatedNode

	directiveByLine := make(map[int]*GompherDirective)

	for _, cg := range file.Comments {
		for _, c := range cg.List {
			if !strings.HasPrefix(c.Text, "//gompher ") && c.Text != "//gompher" {
				continue
			}

			line := fset.Position(c.Pos()).Line
			text := strings.TrimPrefix(c.Text, "//gompher")
			text = strings.TrimSpace(text)

			directive, err := parseDirectiveText(text, c.Pos(), line)
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", line, err)
			}

			directiveByLine[line] = directive
		}
	}

	// directives that need no node — just a sync point
	noNodeDirectives := map[DirectiveKind]bool{
		DirBarrier:  true,
		DirTaskwait: true,
	}

	// track which directives got matched to a node
	matched := make(map[int]bool)

	ast.Inspect(file, func(n ast.Node) bool {
		if n == nil {
			return false
		}

		block, ok := n.(*ast.BlockStmt)
		if !ok {
			return true
		}

		for _, stmt := range block.List {
			stmtLine := fset.Position(stmt.Pos()).Line

			// look back up to 3 lines to handle empty lines between
			// directive and statement
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
					Directive: dir,
					Node:      stmt,
				})
				break
			}
		}

		return true
	})

	// add directives that need no node
	for line, dir := range directiveByLine {
		if noNodeDirectives[dir.Kind] && !matched[line] {
			result = append(result, AnnotatedNode{
				Directive: dir,
				Node:      nil,
			})
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Directive.Line < result[j].Directive.Line
	})

	return result, nil
}

func parseDirectiveText(text string, pos token.Pos, line int) (*GompherDirective, error) {
	if text == "" {
		return nil, fmt.Errorf("empty //gompher directive")
	}

	kind, rest, err := extractKind(text)
	if err != nil {
		return nil, err
	}

	var clauses []Clause
	var subtype string

	switch kind {
	case DirAtomic:
		// "read" | "write" | "update" — stored in Subtype, not Clauses
		if rest != "" {
			valid := map[string]bool{"read": true, "write": true, "update": true}
			if !valid[rest] {
				return nil, fmt.Errorf("invalid atomic type: %q, expected read, write or update", rest)
			}
			subtype = rest
		}

	case DirCritical:
		// optional lock name — stored in Subtype, not Clauses
		if rest != "" {
			name := strings.Trim(rest, "()")
			if name == "" {
				return nil, fmt.Errorf("critical name cannot be empty")
			}
			subtype = name
		}

	default:
		clauses, err = extractClauses(rest)
		if err != nil {
			return nil, err
		}
	}

	if err := validateClauses(kind, clauses); err != nil {
		return nil, err
	}

	return &GompherDirective{
		Kind:    kind,
		Clauses: clauses,
		Subtype: subtype,
		Node:    ast.Node(nil),
		Pos:     pos,
		Line:    line,
	}, nil
}

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

// validClauses defines which clause kinds are legal for each directive.
// directives not in this map accept no clauses.
var validClauses = map[DirectiveKind][]ClauseKind{
	DirParallel:    {ClausePrivate, ClauseFirstPrivate, ClauseShared},
	DirFor:         {ClausePrivate, ClauseFirstPrivate, ClauseSchedule},
	DirParallelFor: {ClausePrivate, ClauseFirstPrivate, ClauseLastPrivate, ClauseShared, ClauseReduction, ClauseSchedule},
	DirSections:    {ClausePrivate, ClauseFirstPrivate, ClauseLastPrivate, ClauseReduction},
	DirSingle:      {ClausePrivate, ClauseFirstPrivate},
	DirTask:        {ClausePrivate, ClauseFirstPrivate},
	DirTaskloop:    {ClausePrivate, ClauseFirstPrivate, ClauseGrainsize},
	// these accept no clauses — not in map
	// DirSection, DirMaster, DirBarrier, DirAtomic, DirCritical, DirTaskwait, DirTaskgroup
}

func validateClauses(kind DirectiveKind, clauses []Clause) error {
	allowed, exists := validClauses[kind]

	// directives not in the map accept no clauses at all
	if !exists {
		if len(clauses) > 0 {
			return fmt.Errorf("directive %q accepts no clauses", kind)
		}
		return nil
	}

	// check each clause is in the allowed list
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
