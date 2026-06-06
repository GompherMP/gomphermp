package transformer

import (
	"fmt"
	"go/ast"

	"github.com/gomphermp/gomphermp/internal/parser"
)

// transformTask rewrites a //gompher task directive.
// Without depend clauses: runtime.Task(func() { body })
// With depend clauses:    runtime.TaskWithDepend(func() { body }, ins, outs, inouts)
func transformTask(result *parser.ParseResult, d parser.TaskDirective) error {
	var ins, outs, inouts []string
	for _, clause := range d.Clauses {
		if dep, ok := clause.(parser.DependClause); ok {
			switch dep.DepType {
			case "in":
				ins = append(ins, dep.Vars...)
			case "out":
				outs = append(outs, dep.Vars...)
			case "inout":
				inouts = append(inouts, dep.Vars...)
			}
		}
	}

	if len(ins) == 0 && len(outs) == 0 && len(inouts) == 0 {
		return transformBlockDirective(result, d.Node, d.Pos, d.Line, "Task")
	}

	// depend path: TaskWithDepend(func() { body }, ins, outs, inouts)
	// Cannot use transformBlockDirective here because TaskWithDepend's closure
	// is the first argument, not the last.
	body, ok := d.Node.(*ast.BlockStmt)
	if !ok {
		return fmt.Errorf("Task at line %d: expected *ast.BlockStmt, got %T", d.Line, d.Node)
	}

	closure := buildClosure(body)
	call := buildRuntimeCall("TaskWithDepend",
		closure,
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
