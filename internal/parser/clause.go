package parser

// ClauseKind identifies the type of a data-sharing clause.
type ClauseKind string

const (
	ClausePrivate      ClauseKind = "private"
	ClauseFirstPrivate ClauseKind = "firstprivate"
	ClauseLastPrivate  ClauseKind = "lastprivate"
	ClauseShared       ClauseKind = "shared"
	ClauseReduction    ClauseKind = "reduction"
	ClauseSchedule     ClauseKind = "schedule"
	ClauseDepend       ClauseKind = "depend"
	ClauseGrainsize    ClauseKind = "grainsize"
)

// The unexported method clauseKind() follows the same pattern as go/ast.
// Only types defined inside this package can implement Clause.
// In the transformer, clauses are accessed via type switch:
type Clause interface {
	clauseKind() ClauseKind
}

// PrivateClause represents private(x, y, z).
// Each goroutine receives its own uninitialized copy of each variable in Vars.
type PrivateClause struct{ Vars []string }

func (c PrivateClause) clauseKind() ClauseKind { return ClausePrivate }

// FirstPrivateClause represents firstprivate(x, y).
// Like PrivateClause but each copy is initialized with the original value
// before the goroutine enters the parallel region or task.
type FirstPrivateClause struct{ Vars []string }

func (c FirstPrivateClause) clauseKind() ClauseKind { return ClauseFirstPrivate }

// LastPrivateClause represents lastprivate(x).
// Like PrivateClause but the copy belonging to the last sequential iteration
// or lexically last section is written back to the original variable when done.
type LastPrivateClause struct{ Vars []string }

func (c LastPrivateClause) clauseKind() ClauseKind { return ClauseLastPrivate }

// SharedClause represents shared(x, y).
// All goroutines access the same memory location for each variable in Vars.
type SharedClause struct{ Vars []string }

func (c SharedClause) clauseKind() ClauseKind { return ClauseShared }

// ReductionClause represents reduction(op:x, y).
// Creates private copies initialized to the operator's neutral value,
// then combines them into the original variable at the end.
type ReductionClause struct {
	Operator string
	Vars     []string
}

func (c ReductionClause) clauseKind() ClauseKind { return ClauseReduction }

// ScheduleClause represents schedule(kind) or schedule(kind, chunk_size).
// Controls how loop iterations are divided and assigned to goroutines.
type ScheduleClause struct {
	Kind  string // "static" | "dynamic"
	Chunk string // chunk size or empty for default
}

func (c ScheduleClause) clauseKind() ClauseKind { return ClauseSchedule }

// DependClause represents depend(type:x, y).
// Used with task directives to declare data dependencies between tasks,
// controlling the order in which tasks are allowed to execute.
type DependClause struct {
	DepType string   // "in" | "out" | "inout"
	Vars    []string // variable names, e.g. ["x", "buffer"]
}

func (c DependClause) clauseKind() ClauseKind { return ClauseDepend }

// GrainsizeClause represents grainsize(n).
// Used with taskloop to control how many iterations are bundled into each task.
type GrainsizeClause struct{ Size string }

func (c GrainsizeClause) clauseKind() ClauseKind { return ClauseGrainsize }
