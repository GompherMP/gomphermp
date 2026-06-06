package transformer

import (
	"fmt"
	"go/ast"

	"github.com/gomphermp/gomphermp/internal/parser"
)

// transformTask rewrites a //gompher task directive.
//
//   - No clauses:                   runtime.Task(func() { body })
//   - depend only:                  runtime.TaskWithDepend(func() { body }, ins, outs, inouts)
//   - firstprivate(x):              _fp_x := x; runtime.Task(func() { x := _fp_x; body })
//   - private(x):                   runtime.Task(func() { var x T; body })
//   - shared(x):                    silently ignored — Go closures share by reference by default
//   - combinations of the above are supported
func transformTask(result *parser.ParseResult, d parser.TaskDirective) error {
	var fpVars, pvVars []string
	var ins, outs, inouts []string

	for _, clause := range d.Clauses {
		switch c := clause.(type) {
		case parser.DependClause:
			switch c.DepType {
			case "in":
				ins = append(ins, c.Vars...)
			case "out":
				outs = append(outs, c.Vars...)
			case "inout":
				inouts = append(inouts, c.Vars...)
			}
		case parser.FirstPrivateClause:
			fpVars = append(fpVars, c.Vars...)
		case parser.PrivateClause:
			pvVars = append(pvVars, c.Vars...)
		// SharedClause: silently ignored
		}
	}

	hasDep := len(ins) > 0 || len(outs) > 0 || len(inouts) > 0
	hasFP := len(fpVars) > 0
	hasPV := len(pvVars) > 0

	// Fast paths when no private/firstprivate clauses are present.
	if !hasFP && !hasPV {
		if !hasDep {
			return transformBlockDirective(result, d.Node, d.Pos, d.Line, "Task")
		}
		// depend-only: TaskWithDepend(func() { body }, ins, outs, inouts)
		// Closure is the first argument, unlike other directives, so we cannot
		// use transformBlockDirective here.
		body, ok := d.Node.(*ast.BlockStmt)
		if !ok {
			return fmt.Errorf("Task at line %d: expected *ast.BlockStmt, got %T", d.Line, d.Node)
		}
		call := buildRuntimeCall("TaskWithDepend",
			buildClosure(body),
			buildUintptrSlice(ins),
			buildUintptrSlice(outs),
			buildUintptrSlice(inouts),
		)
		if !replaceBlockStmt(result.File, body, call) {
			return fmt.Errorf("Task at line %d: body block not found in AST", d.Line)
		}
		removeDirectiveComment(result.File, d.Pos)
		ensureUnsafeImport(result.File)
		return nil
	}

	// private/firstprivate path.
	body, ok := d.Node.(*ast.BlockStmt)
	if !ok {
		return fmt.Errorf("Task at line %d: expected *ast.BlockStmt, got %T", d.Line, d.Node)
	}

	// Build statements to prepend inside the closure body.
	var closurePrefix []ast.Stmt

	// private(x): var x T — zero-initialized, shadows outer x.
	for _, v := range pvVars {
		typeExpr, err := findVarType(result.File, v, d.Pos)
		if err != nil {
			return fmt.Errorf("Task at line %d: %w", d.Line, err)
		}
		closurePrefix = append(closurePrefix, buildPrivateVarDecl(v, typeExpr))
	}

	// firstprivate(x): x := _fp_x — shadows outer x with the captured copy.
	if hasFP {
		closurePrefix = append(closurePrefix, buildFirstprivateShadow(fpVars))
	}

	modifiedBody := &ast.BlockStmt{List: append(closurePrefix, body.List...)}
	closure := buildClosure(modifiedBody)

	var call ast.Stmt
	if hasDep {
		call = buildRuntimeCall("TaskWithDepend",
			closure,
			buildUintptrSlice(ins),
			buildUintptrSlice(outs),
			buildUintptrSlice(inouts),
		)
	} else {
		call = buildRuntimeCall("Task", closure)
	}

	if hasFP {
		// Inject the capture assignment (_fp_x := x) immediately before the
		// Task call in the surrounding block, then replace body with the call.
		captureStmt := buildFirstprivateCapture(fpVars)
		if !replaceBlockStmtWithPrefix(result.File, body, call, []ast.Stmt{captureStmt}) {
			return fmt.Errorf("Task at line %d: body block not found in AST", d.Line)
		}
	} else {
		if !replaceBlockStmt(result.File, body, call) {
			return fmt.Errorf("Task at line %d: body block not found in AST", d.Line)
		}
	}

	removeDirectiveComment(result.File, d.Pos)
	if hasDep {
		ensureUnsafeImport(result.File)
	}
	return nil
}

// transformTaskwait rewrites a //gompher taskwait directive into runtime.Taskwait().
// Unlike block directives, taskwait has no associated AST node — it is a standalone
// sync point. The call is spliced into the parent block at the directive's position.
func transformTaskwait(result *parser.ParseResult, d parser.TaskwaitDirective) error {
	call := buildRuntimeCall("Taskwait")
	if !insertCallAtPos(result.File, d.Pos, call) {
		return fmt.Errorf("Taskwait at line %d: could not find containing block", d.Line)
	}
	removeDirectiveComment(result.File, d.Pos)
	return nil
}

// transformTaskgroup rewrites a //gompher taskgroup directive into
// runtime.Taskgroup(func() { body }). Taskgroup waits for all descendant
// tasks in its scope to complete, unlike Taskwait which only waits for
// direct children.
func transformTaskgroup(result *parser.ParseResult, d parser.TaskgroupDirective) error {
	return transformBlockDirective(result, d.Node, d.Pos, d.Line, "Taskgroup")
}
