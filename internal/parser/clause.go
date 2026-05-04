package parser

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
	ClauseAtomicType   ClauseKind = "atomictype"
	ClauseCriticalName ClauseKind = "criticalname"
)

type Clause interface {
	clauseKind() ClauseKind
}

type PrivateClause struct{ Vars []string }

func (c PrivateClause) clauseKind() ClauseKind { return ClausePrivate }

type FirstPrivateClause struct{ Vars []string }

func (c FirstPrivateClause) clauseKind() ClauseKind { return ClauseFirstPrivate }

type LastPrivateClause struct{ Vars []string }

func (c LastPrivateClause) clauseKind() ClauseKind { return ClauseLastPrivate }

type SharedClause struct{ Vars []string }

func (c SharedClause) clauseKind() ClauseKind { return ClauseShared }

type ReductionClause struct {
	Operator string
	Vars     []string
}

func (c ReductionClause) clauseKind() ClauseKind { return ClauseReduction }

type ScheduleClause struct {
	Kind  string // "static" | "dynamic"
	Chunk string
}

func (c ScheduleClause) clauseKind() ClauseKind { return ClauseSchedule }

type DependClause struct {
	DepType string // "in" | "out" | "inout"
	Vars    []string
}

func (c DependClause) clauseKind() ClauseKind { return ClauseDepend }

type GrainsizeClause struct{ Size string }

func (c GrainsizeClause) clauseKind() ClauseKind { return ClauseGrainsize }

type AtomicTypeClause struct{ Type string }

func (c AtomicTypeClause) clauseKind() ClauseKind { return ClauseAtomicType }

type CriticalNameClause struct{ Name string }

func (c CriticalNameClause) clauseKind() ClauseKind { return ClauseCriticalName }
