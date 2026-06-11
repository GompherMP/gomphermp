package transformer

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/gomphermp/gomphermp/internal/parser"
)

// transformSections rewrites a //gompher sections block and its nested
// //gompher section blocks into a single runtime.Sections call (the worksharing
// form, distributed across the enclosing parallel team).
//
//	//gompher sections          runtime.Sections([]func(){
//	{                               func() { A },
//	    //gompher section           func() { B },
//	    { A }                   })
//	    //gompher section
//	    { B }
//	}
func transformSections(result *parser.ParseResult, d parser.SectionsDirective) error {
	return transformSectionsConstruct(result, d.Node, d.Pos, d.Line, "Sections")
}

// transformParallelSections rewrites a //gompher parallel sections block into a
// runtime.ParallelSections call: the combined construct that provisions a team
// and distributes the section blocks across it in one step.
func transformParallelSections(result *parser.ParseResult, d parser.ParallelSectionsDirective) error {
	return transformSectionsConstruct(result, d.Node, d.Pos, d.Line, "ParallelSections")
}

// transformSectionsConstruct implements the shared rewrite for sections and
// parallel sections. Each nested section's body becomes one func() closure
// element of the []func() slice passed to runtimeFunc.
func transformSectionsConstruct(result *parser.ParseResult, node ast.Node, dirPos token.Pos, line int, runtimeFunc string) error {
	outer, ok := node.(*ast.BlockStmt)
	if !ok {
		return fmt.Errorf("%s at line %d: expected *ast.BlockStmt, got %T", runtimeFunc, line, node)
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
		return fmt.Errorf("%s at line %d: no section blocks found", runtimeFunc, line)
	}

	call := buildRuntimeCall(runtimeFunc, buildFuncSlice(elems))

	if !replaceBlockStmt(result.File, outer, call) {
		return fmt.Errorf("%s at line %d: body block not found in AST", runtimeFunc, line)
	}

	// Strip the sections comment and every section comment so go/format does
	// not leave them orphaned around the synthesized call.
	removeDirectiveComment(result.File, dirPos)
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
