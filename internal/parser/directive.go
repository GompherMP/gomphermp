package parser

import (
	"go/ast"
	"go/token"
)

type DirectiveKind string

const (
	DirParallel    DirectiveKind = "parallel"
	DirFor         DirectiveKind = "for"
	DirParallelFor DirectiveKind = "parallel for"
	DirSections    DirectiveKind = "sections"
	DirSection     DirectiveKind = "section"
	DirSingle      DirectiveKind = "single"
	DirMaster      DirectiveKind = "master"
	DirCritical    DirectiveKind = "critical"
	DirBarrier     DirectiveKind = "barrier"
	DirAtomic      DirectiveKind = "atomic"
	DirTask        DirectiveKind = "task"
	DirTaskwait    DirectiveKind = "taskwait"
	DirTaskgroup   DirectiveKind = "taskgroup"
	DirTaskloop    DirectiveKind = "taskloop"
)

type GompherDirective struct {
	Kind    DirectiveKind
	Clauses []Clause
	Subtype string
	Node    ast.Node
	Pos     token.Pos
	Line    int
}
