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

	switch kind {
	case DirAtomic:
		if rest != "" {
			valid := map[string]bool{"read": true, "write": true, "update": true}
			if !valid[rest] {
				return nil, fmt.Errorf("invalid atomic type: %q, expected read, write or update", rest)
			}
			clauses = []Clause{AtomicTypeClause{Type: rest}}
		}

	case DirCritical:
		if rest != "" {
			name := strings.Trim(rest, "()")
			if name == "" {
				return nil, fmt.Errorf("critical name cannot be empty")
			}
			clauses = []Clause{CriticalNameClause{Name: name}}
		}

	default:
		clauses, err = extractClauses(rest)
		if err != nil {
			return nil, err
		}
	}

	return &GompherDirective{
		Kind:    kind,
		Clauses: clauses,
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
