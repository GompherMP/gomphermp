package transformer

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/gomphermp/gomphermp/internal/parser"
)

// transformSections rewrites a //gompher sections block and its nested
// //gompher section blocks into a single runtime.Sections call.
//
//	//gompher sections          runtime.Sections([]func(){
//	{                               func() { A },
//	    //gompher section           func() { B },
//	    { A }                   })
//	    //gompher section
//	    { B }
//	}
//
// Each section's body block becomes one func() closure element of the
// []func() slice passed to runtime.Sections, which dispatches the closures
// across the pool. Because the section directives are nested children of the
// sections directive (not independent constructs), they are consumed here:
// Transform does not dispatch SectionDirective on its own, so the only place
// a section is ever rewritten is from its enclosing sections block.
func transformSections(result *parser.ParseResult, d parser.SectionsDirective) error {
	outer, ok := d.Node.(*ast.BlockStmt)
	if !ok {
		return fmt.Errorf("Sections at line %d: expected *ast.BlockStmt, got %T", d.Line, d.Node)
	}

	// Walk the outer block in source order, picking out the inner blocks that
	// the parser tagged as sections. Matching by pointer against the parsed
	// SectionDirective nodes guarantees we only wrap genuine section blocks
	// (and, for nested sections, only those belonging to this construct).
	var elems []ast.Expr
	var sectionPositions []token.Pos
	for _, stmt := range outer.List {
		block, ok := stmt.(*ast.BlockStmt)
		if !ok {
			continue
		}
		sd, found := findSectionDirective(result, block)
		if !found {
			continue
		}
		elems = append(elems, buildClosure(block))
		sectionPositions = append(sectionPositions, sd.Pos)
	}

	if len(elems) == 0 {
		return fmt.Errorf("Sections at line %d: no section blocks found", d.Line)
	}

	call := buildRuntimeCall("Sections", buildFuncSlice(elems))

	if !replaceBlockStmt(result.File, outer, call) {
		return fmt.Errorf("Sections at line %d: body block not found in AST", d.Line)
	}

	// Strip the sections comment and every section comment so go/format does
	// not leave them orphaned around the synthesized call.
	removeDirectiveComment(result.File, d.Pos)
	for _, pos := range sectionPositions {
		removeDirectiveComment(result.File, pos)
	}

	return nil
}

// findSectionDirective returns the SectionDirective whose body is block, if one
// exists among the parsed nodes. The pointer comparison ties a raw *ast.BlockStmt
// back to the section the parser recognized for it.
func findSectionDirective(result *parser.ParseResult, block *ast.BlockStmt) (parser.SectionDirective, bool) {
	for _, node := range result.Nodes {
		if sd, ok := node.Directive.(parser.SectionDirective); ok && sd.Node == block {
			return sd, true
		}
	}
	return parser.SectionDirective{}, false
}
