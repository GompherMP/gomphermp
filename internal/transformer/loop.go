package transformer

import (
	"fmt"
	"go/ast"
	"go/token"
	"github.com/gomphermp/gomphermp/internal/parser"
)

// transformFor rewrites a //gompher for directive into a worksharing call over
// the current team: runtime.For(threadID, func(i int){...}, N). It is meant to
// appear inside a //gompher parallel region, where threadID names the enclosing
// closure's parameter; the threadID argument is what lets each team goroutine
// claim its own static chunk.
func transformFor(result *parser.ParseResult, d parser.ForDirective) error {
	return transformLoopDirective(result, d.Node, d.Pos, d.Line, d.Clauses, "For", true)
}

// transformParallelFor rewrites a //gompher parallel for directive, the
// combined construct that creates a team and distributes the loop in one call.
// runtime.ParallelFor provisions the team and wires the per-goroutine threadID
// internally, so the emitted call takes no threadID argument.
func transformParallelFor(result *parser.ParseResult, d parser.ParallelForDirective) error {
	return transformLoopDirective(result, d.Node, d.Pos, d.Line, d.Clauses, "ParallelFor", false)
}

// transformLoopDirective is the shared rewrite for the loop directives (for,
// parallel for): it analyzes the canonical for, wraps the body in a func(int)
// closure over the normalized space [0, count), and replaces the for statement.
// The simple form `for v := 0; v < N; v++` uses v as the parameter and N as the
// count; any other canonical form is normalized to a 0-based counter, with the
// induction variable recovered via `v := lb ± counter*step` (as OpenMP lowers
// loops).
//
// staticFunc is the runtime entry point for static/absent scheduling ("For" or
// "ParallelFor"); a schedule clause redirects to the *Dynamic or *StaticChunked
// variant. prependThreadID is true for the worksharing For. Data-sharing clauses
// route through transformLoopWithClauses.
func transformLoopDirective(
	result *parser.ParseResult,
	node ast.Node,
	dirPos token.Pos,
	line int,
	clauses []parser.Clause,
	staticFunc string,
	prependThreadID bool,
) error {
	forStmt, ok := node.(*ast.ForStmt)
	if !ok {
		return fmt.Errorf("%s at line %d: expected *ast.ForStmt, got %T", staticFunc, line, node)
	}

	form, err := analyzeLoop(forStmt)
	if err != nil {
		return fmt.Errorf("%s at line %d: %w", staticFunc, line, err)
	}

	// The runtime distributes the normalized index space [0, count). For the
	// simple canonical form that is just [0, bound); otherwise the body is
	// driven by a 0-based counter and the user's induction variable is
	// recovered at the top of the body.
	count := form.iterationCount()
	counterName := form.loopVar
	if !form.simple() {
		counterName = loopCounterName
		forStmt.Body.List = append([]ast.Stmt{form.inductionRecovery()}, forStmt.Body.List...)
	}

	closurePrefix, capturePrefix, err := dataClausePrefixes(result, clauses, dirPos)
	if err != nil {
		return fmt.Errorf("%s at line %d: %w", staticFunc, line, err)
	}
	rvars, err := reductionVars(result, clauses, dirPos)
	if err != nil {
		return fmt.Errorf("%s at line %d: %w", staticFunc, line, err)
	}
	lpvars, err := lastprivateVars(result, clauses, dirPos)
	if err != nil {
		return fmt.Errorf("%s at line %d: %w", staticFunc, line, err)
	}

	// lastprivate appends a writeback to the loop body before it is wrapped in a
	// closure, so the goroutine running the last iteration (counter == count-1)
	// copies its value out.
	for _, lp := range lpvars {
		forStmt.Body.List = append(forStmt.Body.List, buildLastprivateWriteback(lp, counterName, count))
	}
	closure := buildClosureWithIntParam(forStmt.Body, counterName)

	if len(closurePrefix) == 0 && len(capturePrefix) == 0 && len(rvars) == 0 && len(lpvars) == 0 {
		// No data clauses: the simple, direct rewrite.
		call := buildLoopCall(clauses, staticFunc, prependThreadID, closure, count)
		if !replaceForStmt(result.File, forStmt, call) {
			return fmt.Errorf("%s at line %d: for statement not found in AST", staticFunc, line)
		}
		removeDirectiveComment(result.File, dirPos)
		return nil
	}

	return transformLoopWithClauses(result, forStmt, dirPos, line, staticFunc, prependThreadID, clauses, closure, count, closurePrefix, capturePrefix, rvars, lpvars)
}

// transformLoopWithClauses expands a clause-carrying loop so the per-goroutine
// copies (private/firstprivate shadows, reduction accumulators) live in a scope
// that runs exactly once per goroutine:
//
//   - bare for: a plain block wraps the worksharing call. The block already
//     runs once per goroutine because it sits inside the enclosing parallel.
//   - parallel for: a runtime.Parallel(func(threadID int){...}) wraps the
//     worksharing call, providing that per-goroutine scope itself.
//
// firstprivate captures and reduction pointer captures are spliced in just
// before the wrapper (they read the outer variables before they are shadowed);
// reduction combines run after the loop, folding each goroutine's partial back.
func transformLoopWithClauses(
	result *parser.ParseResult,
	forStmt *ast.ForStmt,
	dirPos token.Pos,
	line int,
	staticFunc string,
	prependThreadID bool,
	clauses []parser.Clause,
	closure ast.Expr,
	bound ast.Expr,
	closurePrefix, capturePrefix []ast.Stmt,
	rvars []reductionVar,
	lpvars []lastprivateVar,
) error {
	// Inside a clause expansion the loop is always the worksharing form
	// (For/ForDynamic), since the team is provided by the surrounding block or
	// the synthesized Parallel.
	inner := buildWorksharingLoopCall(clauses, closure, bound)

	// Per-goroutine body, shared by both shapes.
	var body []ast.Stmt
	body = append(body, closurePrefix...)
	body = append(body, buildLastprivateDecls(lpvars)...)
	body = append(body, buildReductionAccumulators(rvars)...)
	body = append(body, inner)
	body = append(body, buildReductionCombines(rvars)...)

	// Outer captures that must precede the wrapper.
	var pre []ast.Stmt
	pre = append(pre, capturePrefix...)
	pre = append(pre, buildLastprivateCaptures(lpvars)...)
	pre = append(pre, buildReductionCaptures(rvars)...)

	var ok bool
	if prependThreadID {
		// bare for: the block itself is the per-goroutine scope, so the captures
		// go inside it (still before the shadows).
		blockList := append(pre, body...)
		ok = replaceForStmt(result.File, forStmt, &ast.BlockStmt{List: blockList})
	} else {
		// parallel for: wrap the body in a Parallel closure; captures precede it.
		parClosure := buildClosureWithIntParam(&ast.BlockStmt{List: body}, threadIDParamName)
		parCall := buildRuntimeCall("Parallel", parClosure)
		ok = replaceStmtWithPrefix(result.File, forStmt, parCall, pre)
	}
	if !ok {
		return fmt.Errorf("%s at line %d: for statement not found in AST", staticFunc, line)
	}
	removeDirectiveComment(result.File, dirPos)
	return nil
}

// buildLoopCall builds the direct (clause-free) loop call. Scheduling chooses the
// runtime entry point: *Dynamic for schedule(dynamic[, chunk]), *StaticChunked
// for schedule(static, chunk), and the plain block-static For/ParallelFor for
// static without a chunk (or no schedule clause).
func buildLoopCall(clauses []parser.Clause, staticFunc string, prependThreadID bool, closure, bound ast.Expr) ast.Stmt {
	if sched, ok := findSchedule(clauses); ok {
		switch {
		case sched.Kind == "dynamic":
			return buildRuntimeCall(staticFunc+"Dynamic", closure, bound, scheduleChunkLit(sched))
		case sched.Kind == "static" && sched.Chunk != "":
			if prependThreadID {
				return buildRuntimeCall(staticFunc+"StaticChunked", &ast.Ident{Name: threadIDParamName}, closure, bound, scheduleChunkLit(sched))
			}
			return buildRuntimeCall(staticFunc+"StaticChunked", closure, bound, scheduleChunkLit(sched))
		}
	}
	if prependThreadID {
		return buildRuntimeCall(staticFunc, &ast.Ident{Name: threadIDParamName}, closure, bound)
	}
	return buildRuntimeCall(staticFunc, closure, bound)
}

// buildWorksharingLoopCall builds the worksharing loop call used inside a clause
// expansion: runtime.For / ForDynamic / ForStaticChunked depending on schedule.
// It never emits a combined Parallel* form, because the team is already
// established by the enclosing scope.
func buildWorksharingLoopCall(clauses []parser.Clause, closure, bound ast.Expr) ast.Stmt {
	if sched, ok := findSchedule(clauses); ok {
		switch {
		case sched.Kind == "dynamic":
			return buildRuntimeCall("ForDynamic", closure, bound, scheduleChunkLit(sched))
		case sched.Kind == "static" && sched.Chunk != "":
			return buildRuntimeCall("ForStaticChunked", &ast.Ident{Name: threadIDParamName}, closure, bound, scheduleChunkLit(sched))
		}
	}
	return buildRuntimeCall("For", &ast.Ident{Name: threadIDParamName}, closure, bound)
}

// scheduleChunkLit returns the chunk-size literal for a dynamic schedule,
// defaulting to 1 when none was given.
func scheduleChunkLit(sched parser.ScheduleClause) *ast.BasicLit {
	chunk := sched.Chunk
	if chunk == "" {
		chunk = "1"
	}
	return &ast.BasicLit{Kind: token.INT, Value: chunk}
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
