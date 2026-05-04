package parser

import (
	"testing"
)

// =============================================
// LEVEL 1: extractClauses() in isolation
// =============================================

func TestExtractClauses_Private(t *testing.T) {
	clauses, err := extractClauses("private(x, y)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(clauses) != 1 {
		t.Fatalf("expected 1 clause, got %d", len(clauses))
	}
	c, ok := clauses[0].(PrivateClause)
	if !ok {
		t.Fatalf("expected PrivateClause, got %T", clauses[0])
	}
	if len(c.Vars) != 2 || c.Vars[0] != "x" || c.Vars[1] != "y" {
		t.Errorf("incorrect vars: %v", c.Vars)
	}
}

func TestExtractClauses_FirstPrivate(t *testing.T) {
	clauses, err := extractClauses("firstprivate(base)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c, ok := clauses[0].(FirstPrivateClause)
	if !ok {
		t.Fatalf("expected FirstPrivateClause, got %T", clauses[0])
	}
	if len(c.Vars) != 1 || c.Vars[0] != "base" {
		t.Errorf("incorrect vars: %v", c.Vars)
	}
}

func TestExtractClauses_LastPrivate(t *testing.T) {
	clauses, err := extractClauses("lastprivate(x)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c, ok := clauses[0].(LastPrivateClause)
	if !ok {
		t.Fatalf("expected LastPrivateClause, got %T", clauses[0])
	}
	if len(c.Vars) != 1 || c.Vars[0] != "x" {
		t.Errorf("incorrect vars: %v", c.Vars)
	}
}

func TestExtractClauses_Shared(t *testing.T) {
	clauses, err := extractClauses("shared(global, contador)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c, ok := clauses[0].(SharedClause)
	if !ok {
		t.Fatalf("expected SharedClause, got %T", clauses[0])
	}
	if len(c.Vars) != 2 {
		t.Errorf("incorrect vars: %v", c.Vars)
	}
}

func TestExtractClauses_Reduction_Sum(t *testing.T) {
	clauses, err := extractClauses("reduction(+:suma)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c, ok := clauses[0].(ReductionClause)
	if !ok {
		t.Fatalf("expected ReductionClause, got %T", clauses[0])
	}
	if c.Operator != "+" {
		t.Errorf("incorrect operator: %q", c.Operator)
	}
	if len(c.Vars) != 1 || c.Vars[0] != "suma" {
		t.Errorf("incorrect vars: %v", c.Vars)
	}
}

func TestExtractClauses_Reduction_And(t *testing.T) {
	clauses, err := extractClauses("reduction(&&:result)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c, ok := clauses[0].(ReductionClause)
	if !ok {
		t.Fatalf("expected ReductionClause, got %T", clauses[0])
	}
	if c.Operator != "&&" {
		t.Errorf("incorrect operator: %q", c.Operator)
	}
}

func TestExtractClauses_Reduction_Max(t *testing.T) {
	clauses, err := extractClauses("reduction(max:result)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c, ok := clauses[0].(ReductionClause)
	if !ok {
		t.Fatalf("expected ReductionClause, got %T", clauses[0])
	}
	if c.Operator != "max" {
		t.Errorf("incorrect operator: %q", c.Operator)
	}
}

func TestExtractClauses_ScheduleStatic(t *testing.T) {
	clauses, err := extractClauses("schedule(static, 10)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c, ok := clauses[0].(ScheduleClause)
	if !ok {
		t.Fatalf("expected ScheduleClause, got %T", clauses[0])
	}
	if c.Kind != "static" {
		t.Errorf("incorrect kind: %q", c.Kind)
	}
	if c.Chunk != "10" {
		t.Errorf("incorrect chunk: %q", c.Chunk)
	}
}

func TestExtractClauses_ScheduleStaticNoChunk(t *testing.T) {
	clauses, err := extractClauses("schedule(static)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c, ok := clauses[0].(ScheduleClause)
	if !ok {
		t.Fatalf("expected ScheduleClause, got %T", clauses[0])
	}
	if c.Kind != "static" {
		t.Errorf("incorrect kind: %q", c.Kind)
	}
	if c.Chunk != "" {
		t.Errorf("chunk should be empty, got %q", c.Chunk)
	}
}

func TestExtractClauses_ScheduleDynamic(t *testing.T) {
	clauses, err := extractClauses("schedule(dynamic, 5)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c, ok := clauses[0].(ScheduleClause)
	if !ok {
		t.Fatalf("expected ScheduleClause, got %T", clauses[0])
	}
	if c.Kind != "dynamic" {
		t.Errorf("incorrect kind: %q", c.Kind)
	}
	if c.Chunk != "5" {
		t.Errorf("incorrect chunk: %q", c.Chunk)
	}
}

func TestExtractClauses_ScheduleDynamicNoChunk(t *testing.T) {
	clauses, err := extractClauses("schedule(dynamic)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c, ok := clauses[0].(ScheduleClause)
	if !ok {
		t.Fatalf("expected ScheduleClause, got %T", clauses[0])
	}
	if c.Chunk != "" {
		t.Errorf("chunk should be empty, got %q", c.Chunk)
	}
}

func TestExtractClauses_Multiple(t *testing.T) {
	clauses, err := extractClauses("schedule(static, 10) private(x, y)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(clauses) != 2 {
		t.Fatalf("expected 2 clauses, got %d", len(clauses))
	}
	if _, ok := clauses[0].(ScheduleClause); !ok {
		t.Errorf("first clause should be ScheduleClause, got %T", clauses[0])
	}
	if _, ok := clauses[1].(PrivateClause); !ok {
		t.Errorf("second clause should be PrivateClause, got %T", clauses[1])
	}
}

func TestExtractClauses_Empty(t *testing.T) {
	clauses, err := extractClauses("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(clauses) != 0 {
		t.Errorf("expected 0 clauses, got %d", len(clauses))
	}
}

func TestExtractClauses_Unknown(t *testing.T) {
	_, err := extractClauses("unknownclause(x)")
	if err == nil {
		t.Fatal("expected error for unknown clause")
	}
}

// =============================================
// LEVEL 2: parseDirectiveText() combined
// =============================================

func TestParseDirectiveText_ParallelFor(t *testing.T) {
	dir, err := parseDirectiveText("parallel for schedule(static, 10) private(x)", 0, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dir.Kind != DirParallelFor {
		t.Errorf("incorrect kind: %q", dir.Kind)
	}
	if len(dir.Clauses) != 2 {
		t.Errorf("expected 2 clauses, got %d", len(dir.Clauses))
	}
}

func TestParseDirectiveText_Parallel(t *testing.T) {
	dir, err := parseDirectiveText("parallel shared(global) private(local)", 0, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dir.Kind != DirParallel {
		t.Errorf("incorrect kind: %q", dir.Kind)
	}
	if len(dir.Clauses) != 2 {
		t.Errorf("expected 2 clauses, got %d", len(dir.Clauses))
	}
}

func TestParseDirectiveText_Barrier(t *testing.T) {
	dir, err := parseDirectiveText("barrier", 0, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dir.Kind != DirBarrier {
		t.Errorf("incorrect kind: %q", dir.Kind)
	}
	if len(dir.Clauses) != 0 {
		t.Errorf("barrier should have no clauses, got %d", len(dir.Clauses))
	}
}

func TestParseDirectiveText_AtomicUpdate(t *testing.T) {
	dir, err := parseDirectiveText("atomic update", 0, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dir.Kind != DirAtomic {
		t.Errorf("incorrect kind: %q", dir.Kind)
	}
	c, ok := dir.Clauses[0].(AtomicTypeClause)
	if !ok {
		t.Fatalf("expected AtomicTypeClause, got %T", dir.Clauses[0])
	}
	if c.Type != "update" {
		t.Errorf("incorrect type: %q", c.Type)
	}
}

func TestParseDirectiveText_AtomicRead(t *testing.T) {
	dir, err := parseDirectiveText("atomic read", 0, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c, ok := dir.Clauses[0].(AtomicTypeClause)
	if !ok {
		t.Fatalf("expected AtomicTypeClause, got %T", dir.Clauses[0])
	}
	if c.Type != "read" {
		t.Errorf("incorrect type: %q", c.Type)
	}
}

func TestParseDirectiveText_CriticalNamed(t *testing.T) {
	dir, err := parseDirectiveText("critical(mylock)", 0, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dir.Kind != DirCritical {
		t.Errorf("incorrect kind: %q", dir.Kind)
	}
	c, ok := dir.Clauses[0].(CriticalNameClause)
	if !ok {
		t.Fatalf("expected CriticalNameClause, got %T", dir.Clauses[0])
	}
	if c.Name != "mylock" {
		t.Errorf("incorrect name: %q", c.Name)
	}
}

func TestParseDirectiveText_CriticalAnonymous(t *testing.T) {
	dir, err := parseDirectiveText("critical", 0, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dir.Kind != DirCritical {
		t.Errorf("incorrect kind: %q", dir.Kind)
	}
	if len(dir.Clauses) != 0 {
		t.Errorf("anonymous critical should have no clauses, got %d", len(dir.Clauses))
	}
}

func TestParseDirectiveText_Unknown(t *testing.T) {
	_, err := parseDirectiveText("unknowndirective", 0, 1)
	if err == nil {
		t.Fatal("expected error for unknown directive")
	}
}

// =============================================
// LEVEL 3: Parse() with real Go source code
// =============================================

func TestParse_ParallelFor(t *testing.T) {
	src := `package main

func main() {
	//gompher parallel for schedule(static, 10)
	for i := 0; i < 100; i++ {
	}
}`
	result, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(result.Nodes))
	}
	if result.Nodes[0].Directive.Kind != DirParallelFor {
		t.Errorf("incorrect kind: %q", result.Nodes[0].Directive.Kind)
	}
}

func TestParse_ParallelBlock(t *testing.T) {
	src := `package main

func main() {
	//gompher parallel shared(global) private(local)
	{
		local = 1
		global = global + local
	}
}`
	result, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(result.Nodes))
	}
	if result.Nodes[0].Directive.Kind != DirParallel {
		t.Errorf("incorrect kind: %q", result.Nodes[0].Directive.Kind)
	}
}

func TestParse_MultipleDirectives(t *testing.T) {
	src := `package main

func main() {
	//gompher parallel for private(x)
	for i := 0; i < 100; i++ {
	}

	//gompher barrier

	//gompher parallel reduction(+:suma)
	{
		suma += calcular()
	}
}`
	result, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(result.Nodes))
	}
	if result.Nodes[0].Directive.Kind != DirParallelFor {
		t.Errorf("node 0: incorrect kind: %q", result.Nodes[0].Directive.Kind)
	}
	if result.Nodes[1].Directive.Kind != DirBarrier {
		t.Errorf("node 1: incorrect kind: %q", result.Nodes[1].Directive.Kind)
	}
	if result.Nodes[2].Directive.Kind != DirParallel {
		t.Errorf("node 2: incorrect kind: %q", result.Nodes[2].Directive.Kind)
	}
}

func TestParse_InvalidGoSyntax(t *testing.T) {
	src := `package main

func main() {
	this is not valid go
}`
	_, err := Parse(src)
	if err == nil {
		t.Fatal("expected error for invalid Go syntax")
	}
}
