package transformer

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/gomphermp/gomphermp/internal/parser"
)

// Validate enforces the contextual rule: every worksharing or
// synchronization construct that operates on a team (for, sections, single,
// master, barrier) must appear inside a //gompher parallel region,
// which is what provides the team (and the threadID that for and master reference).
// The combined constructs (parallel for, parallel sections) and the context-free ones
// (critical, atomic) are exempt because they either provision their 
// own team or need none. It runs as a separate pass before Transform so that
// violations are reported with a clear "must appear inside a parallel region" message.
func Validate(result *parser.ParseResult) error {
	if result == nil {
		return nil
	}
	// Collect the brace ranges of every //gompher parallel block. A construct
	// is correctly nested when its directive position falls within one of them.
	var parallelBlocks []*ast.BlockStmt
	for _, node := range result.Nodes {
		if d, ok := node.Directive.(parser.ParallelDirective); ok {
			if b, ok := d.Node.(*ast.BlockStmt); ok {
				parallelBlocks = append(parallelBlocks, b)
			}
		}
	}

	for _, node := range result.Nodes {
		kind, pos, line, mustNest := teamConstruct(node.Directive)
		if !mustNest {
			continue
		}
		if !insideAnyBlock(pos, parallelBlocks) {
			return fmt.Errorf("line %d: directive %q must appear inside a //gompher parallel region", line, kind)
		}
	}
	return nil
}

// teamConstruct reports whether d is a worksharing/synchronization construct
// that must be nested inside a parallel region, returning its kind name and
// source position for diagnostics.
func teamConstruct(d parser.Directive) (kind string, pos token.Pos, line int, mustNest bool) {
	switch v := d.(type) {
	case parser.ForDirective:
		return string(parser.DirFor), v.Pos, v.Line, true
	case parser.SectionsDirective:
		return string(parser.DirSections), v.Pos, v.Line, true
	case parser.SingleDirective:
		return string(parser.DirSingle), v.Pos, v.Line, true
	case parser.MasterDirective:
		return string(parser.DirMaster), v.Pos, v.Line, true
	case parser.BarrierDirective:
		return string(parser.DirBarrier), v.Pos, v.Line, true
	}
	return "", 0, 0, false
}

// insideAnyBlock reports whether pos falls strictly between the braces of any
// of the given blocks.
func insideAnyBlock(pos token.Pos, blocks []*ast.BlockStmt) bool {
	for _, b := range blocks {
		if b.Lbrace < pos && pos < b.Rbrace {
			return true
		}
	}
	return false
}
