package parser

import (
	"go/ast"
	"go/token"
)

// Directive is the interface implemented by all GompherMP directives.
// The unexported methods follow the same pattern as Clause.
// Only types in this package can implement Directive.
type Directive interface {
	directiveKind() DirectiveKind
	line() int
}

// DirectiveKind identifies which //gompher directive was written.
type DirectiveKind string

const (
	DirParallel         DirectiveKind = "parallel"
	DirFor              DirectiveKind = "for"
	DirParallelFor      DirectiveKind = "parallel for"
	DirSections         DirectiveKind = "sections"
	DirParallelSections DirectiveKind = "parallel sections"
	DirSection          DirectiveKind = "section"
	DirSingle           DirectiveKind = "single"
	DirMaster           DirectiveKind = "master"
	DirCritical         DirectiveKind = "critical"
	DirBarrier          DirectiveKind = "barrier"
	DirAtomic           DirectiveKind = "atomic"
	DirTask             DirectiveKind = "task"
	DirTaskwait         DirectiveKind = "taskwait"
	DirTaskgroup        DirectiveKind = "taskgroup"
	DirTaskloop         DirectiveKind = "taskloop"
)

// pos holds source position fields shared by all directives.
// Embedded in each concrete directive type.
// Pos is Go's internal token position.
type pos struct {
	Pos  token.Pos
	Line int
}

// ParallelDirective represents //gompher parallel.
// It defines a parallel region, instantiating a team of goroutines to execute the block concurrently.
type ParallelDirective struct {
	Clauses []Clause // private, firstprivate, shared
	Node    ast.Node // *ast.BlockStmt
	pos
}

func (d ParallelDirective) directiveKind() DirectiveKind { return DirParallel }
func (d ParallelDirective) line() int                    { return d.Line }

// ForDirective represents //gompher for.
// It distributes the iterations of a loop among the existing goroutines of the current team.
type ForDirective struct {
	Clauses []Clause // private, firstprivate, schedule
	Node    ast.Node // *ast.ForStmt
	pos
}

func (d ForDirective) directiveKind() DirectiveKind { return DirFor }
func (d ForDirective) line() int                    { return d.Line }

// ParallelForDirective represents //gompher parallel for.
// A combined construct that creates a parallel region and immediately distributes the loop iterations.
type ParallelForDirective struct {
	Clauses []Clause // private, firstprivate, lastprivate, shared, reduction, schedule
	Node    ast.Node // *ast.ForStmt
	pos
}

func (d ParallelForDirective) directiveKind() DirectiveKind { return DirParallelFor }
func (d ParallelForDirective) line() int                    { return d.Line }

// SectionsDirective represents //gompher sections.
// It defines a set of independent blocks of work to be dynamically distributed among the team.
type SectionsDirective struct {
	Clauses []Clause // private, firstprivate, lastprivate, reduction
	Node    ast.Node // *ast.BlockStmt
	pos
}

func (d SectionsDirective) directiveKind() DirectiveKind { return DirSections }
func (d SectionsDirective) line() int                    { return d.Line }

// ParallelSectionsDirective represents //gompher parallel sections.
// A combined construct that creates a parallel region and immediately
// distributes the enclosed section blocks among the team.
type ParallelSectionsDirective struct {
	Clauses []Clause // private, firstprivate, lastprivate, reduction
	Node    ast.Node // *ast.BlockStmt
	pos
}

func (d ParallelSectionsDirective) directiveKind() DirectiveKind { return DirParallelSections }
func (d ParallelSectionsDirective) line() int                    { return d.Line }

// SectionDirective represents //gompher section.
// It marks a single independent block of work within a sections directive.
type SectionDirective struct {
	Node ast.Node // *ast.BlockStmt
	pos
}

func (d SectionDirective) directiveKind() DirectiveKind { return DirSection }
func (d SectionDirective) line() int                    { return d.Line }

// SingleDirective represents //gompher single.
// It ensures the associated block is executed by only one goroutine in the team (includes an implicit barrier).
type SingleDirective struct {
	Clauses []Clause // private, firstprivate
	Node    ast.Node // *ast.BlockStmt
	pos
}

func (d SingleDirective) directiveKind() DirectiveKind { return DirSingle }
func (d SingleDirective) line() int                    { return d.Line }

// MasterDirective represents //gompher master.
// It ensures the block is executed exclusively by the master goroutine (no implicit barrier).
type MasterDirective struct {
	Node ast.Node // *ast.BlockStmt
	pos
}

func (d MasterDirective) directiveKind() DirectiveKind { return DirMaster }
func (d MasterDirective) line() int                    { return d.Line }

// CriticalDirective represents //gompher critical.
// It guarantees mutual exclusion, serializing access to the block to prevent race conditions.
type CriticalDirective struct {
	Name string   // optional lock name, empty means anonymous
	Node ast.Node // *ast.BlockStmt
	pos
}

func (d CriticalDirective) directiveKind() DirectiveKind { return DirCritical }
func (d CriticalDirective) line() int                    { return d.Line }

// BarrierDirective represents //gompher barrier.
// It specifies an explicit synchronization point where all goroutines in the team must wait.
type BarrierDirective struct {
	pos
	// Node is always nil - barrier is a sync point with no associated code
}

func (d BarrierDirective) directiveKind() DirectiveKind { return DirBarrier }
func (d BarrierDirective) line() int                    { return d.Line }

// AtomicDirective represents //gompher atomic.
// It guarantees that a simple memory operation (read, write, or update) is executed atomically.
type AtomicDirective struct {
	Mode string   // "read" | "write" | "update", empty means update (default)
	Node ast.Node // *ast.ExprStmt or *ast.AssignStmt
	pos
}

func (d AtomicDirective) directiveKind() DirectiveKind { return DirAtomic }
func (d AtomicDirective) line() int                    { return d.Line }

// TaskDirective represents //gompher task.
// It defines an explicit, asynchronous unit of work to be processed by a task pool.
type TaskDirective struct {
	Clauses []Clause // private, firstprivate, depend
	Node    ast.Node // *ast.BlockStmt
	pos
}

func (d TaskDirective) directiveKind() DirectiveKind { return DirTask }
func (d TaskDirective) line() int                    { return d.Line }

// TaskwaitDirective represents //gompher taskwait.
// It synchronizes the current task by pausing execution until all its direct child tasks finish.
type TaskwaitDirective struct {
	pos
	// Node is always nil - taskwait is a sync point with no associated code
}

func (d TaskwaitDirective) directiveKind() DirectiveKind { return DirTaskwait }
func (d TaskwaitDirective) line() int                    { return d.Line }

// TaskgroupDirective represents //gompher taskgroup.
// It provides deep synchronization, waiting for all descendant tasks in its scope to complete.
type TaskgroupDirective struct {
	Node ast.Node // *ast.BlockStmt
	pos
}

func (d TaskgroupDirective) directiveKind() DirectiveKind { return DirTaskgroup }
func (d TaskgroupDirective) line() int                    { return d.Line }

// TaskloopDirective represents //gompher taskloop.
// It distributes loop iterations by generating an asynchronous task for each chunk of iterations.
type TaskloopDirective struct {
	Clauses []Clause // private, firstprivate, grainsize
	Node    ast.Node // *ast.ForStmt
	pos
}

func (d TaskloopDirective) directiveKind() DirectiveKind { return DirTaskloop }
func (d TaskloopDirective) line() int                    { return d.Line }
