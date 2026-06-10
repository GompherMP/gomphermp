package transformer

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/gomphermp/gomphermp/internal/parser"
)

// transformFor rewrites a //gompher for directive into a work-sharing runtime
// call over a canonical loop (for i := 0; i < N; i++). 
func transformFor(result *parser.ParseResult, d parser.ForDirective) error {
	return transformLoopDirective(result, d.Node, d.Pos, d.Line, d.Clauses, "For")
}

// transformParallelFor rewrites a //gompher parallel for directive, the
// combined construct that creates a team and distributes the loop in one call.
func transformParallelFor(result *parser.ParseResult, d parser.ParallelForDirective) error {
	return transformLoopDirective(result, d.Node, d.Pos, d.Line, d.Clauses, "ParallelFor")
}

// transformLoopDirective implements the shared rewrite for the loop-based
// directives like for and parallel for. It extracts the iteration variable and
// upper bound from the canonical for statement, wraps the loop body in a
// func(loopVar int) closure, and replaces the for statement with the
// appropriate runtime call.
func transformLoopDirective(
	result *parser.ParseResult,
	node ast.Node,
	dirPos token.Pos,
	line int,
	clauses []parser.Clause,
	staticFunc string,
	// staticFunc names the runtime function used when scheduling is static or
	// absent ("For" or "ParallelFor"). A schedule(dynamic, chunk) clause overrides
	// this with runtime.ForDynamic regardless of staticFunc.
) error {
	forStmt, ok := node.(*ast.ForStmt)
	if !ok {
		return fmt.Errorf("%s at line %d: expected *ast.ForStmt, got %T", staticFunc, line, node)
	}

	loopVar, err := extractLoopVar(forStmt)
	if err != nil {
		return fmt.Errorf("%s at line %d: %w", staticFunc, line, err)
	}

	bound, err := extractUpperBound(forStmt)
	if err != nil {
		return fmt.Errorf("%s at line %d: %w", staticFunc, line, err)
	}

	closure := buildClosureWithIntParam(forStmt.Body, loopVar)

	var call ast.Stmt
	if sched, ok := findSchedule(clauses); ok && sched.Kind == "dynamic" {
		chunk := sched.Chunk
		if chunk == "" {
			chunk = "1"
		}
		chunkLit := &ast.BasicLit{Kind: token.INT, Value: chunk}
		call = buildRuntimeCall("ForDynamic", closure, bound, chunkLit)
	} else {
		call = buildRuntimeCall(staticFunc, closure, bound)
	}

	if !replaceForStmt(result.File, forStmt, call) {
		return fmt.Errorf("%s at line %d: for statement not found in AST", staticFunc, line)
	}
	removeDirectiveComment(result.File, dirPos)

	return nil
}

// findSchedule returns the first schedule clause in the list, if any. Only one
// schedule clause is meaningful per loop; additional ones are ignored.
func findSchedule(clauses []parser.Clause) (parser.ScheduleClause, bool) {
	for _, c := range clauses {
		if s, ok := c.(parser.ScheduleClause); ok {
			return s, true
		}
	}
	return parser.ScheduleClause{}, false
}
