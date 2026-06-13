package transformer

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/gomphermp/gomphermp/internal/parser"
)

// transformSections rewrites a //gompher sections block and its nested //gompher
// section blocks into one runtime.Sections call - one func() closure per section,
// distributed across the enclosing parallel team.
func transformSections(result *parser.ParseResult, d parser.SectionsDirective) error {
	return transformSectionsConstruct(result, d.Node, d.Pos, d.Line, d.Clauses, "Sections")
}

// transformParallelSections rewrites a //gompher parallel sections block into a
// runtime.ParallelSections call: the combined construct that provisions a team
// and distributes the section blocks across it in one step.
func transformParallelSections(result *parser.ParseResult, d parser.ParallelSectionsDirective) error {
	return transformSectionsConstruct(result, d.Node, d.Pos, d.Line, d.Clauses, "ParallelSections")
}

// transformSectionsConstruct is the shared rewrite for sections and parallel
// sections: each section body becomes one func() element of the []func() passed
// to runtimeFunc. Data-sharing clauses are applied per section: private/
// firstprivate shadow the variable in each closure (firstprivate captured once
// before the call); reduction folds a per-section private copy back under a
// critical section; lastprivate is private in every section and written back
// from the lexically last one.
func transformSectionsConstruct(result *parser.ParseResult, node ast.Node, dirPos token.Pos, line int, clauses []parser.Clause, runtimeFunc string) error {
	outer, ok := node.(*ast.BlockStmt)
	if !ok {
		return fmt.Errorf("%s at line %d: expected *ast.BlockStmt, got %T", runtimeFunc, line, node)
	}

	cd, err := gatherSectionClauses(result, clauses, dirPos)
	if err != nil {
		return fmt.Errorf("%s at line %d: %w", runtimeFunc, line, err)
	}

	// Collect the inner blocks the parser tagged as sections, in source order.
	// Matching by pointer against the parsed SectionDirective nodes ensures we
	// only wrap genuine sections belonging to this construct.
	var blocks []*ast.BlockStmt
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
		blocks = append(blocks, block)
		sectionPositions = append(sectionPositions, sd.Pos)
	}

	if len(blocks) == 0 {
		return fmt.Errorf("%s at line %d: no section blocks found", runtimeFunc, line)
	}

	elems := make([]ast.Expr, len(blocks))
	for i, block := range blocks {
		body := block
		if !cd.empty() {
			isLast := i == len(blocks)-1
			stmts := append([]ast.Stmt{}, cd.sectionPrefix()...)
			stmts = append(stmts, block.List...)
			stmts = append(stmts, cd.sectionSuffix(isLast)...)
			body = &ast.BlockStmt{List: stmts}
		}
		elems[i] = buildClosure(body)
	}

	call := buildRuntimeCall(runtimeFunc, buildFuncSlice(elems))

	captures := cd.captures()
	var replaced bool
	if len(captures) > 0 {
		replaced = replaceBlockStmtWithPrefix(result.File, outer, call, captures)
	} else {
		replaced = replaceBlockStmt(result.File, outer, call)
	}
	if !replaced {
		return fmt.Errorf("%s at line %d: body block not found in AST", runtimeFunc, line)
	}

	// Strip the sections and section comments so go/format does not leave them
	// orphaned around the synthesized call.
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
