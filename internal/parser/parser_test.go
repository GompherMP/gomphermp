package parser

import (
	"strings"
	"testing"
)

// =============================================================================
// Clause parsing - extractClauses() in isolation
// =============================================================================

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

func TestExtractClauses_Depend(t *testing.T) {
	clauses, err := extractClauses("depend(in:x, y)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c, ok := clauses[0].(DependClause)
	if !ok {
		t.Fatalf("expected DependClause, got %T", clauses[0])
	}
	if c.DepType != "in" {
		t.Errorf("incorrect dependency type: %q", c.DepType)
	}
	if len(c.Vars) != 2 || c.Vars[0] != "x" || c.Vars[1] != "y" {
		t.Errorf("incorrect vars: %v", c.Vars)
	}
}

func TestExtractClauses_Depend_In(t *testing.T) {
	clauses, err := extractClauses("depend(in:x)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c, ok := clauses[0].(DependClause)
	if !ok {
		t.Fatalf("expected DependClause, got %T", clauses[0])
	}
	if c.DepType != "in" {
		t.Errorf("incorrect deptype: %q", c.DepType)
	}
	if len(c.Vars) != 1 || c.Vars[0] != "x" {
		t.Errorf("incorrect vars: %v", c.Vars)
	}
}

func TestExtractClauses_Depend_Out(t *testing.T) {
	clauses, err := extractClauses("depend(out:buff)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c, ok := clauses[0].(DependClause)
	if !ok {
		t.Fatalf("expected DependClause, got %T", clauses[0])
	}
	if c.DepType != "out" {
		t.Errorf("incorrect deptype: %q", c.DepType)
	}
}

func TestExtractClauses_Depend_Inout(t *testing.T) {
	clauses, err := extractClauses("depend(inout:buff)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c, ok := clauses[0].(DependClause)
	if !ok {
		t.Fatalf("expected DependClause, got %T", clauses[0])
	}
	if c.DepType != "inout" {
		t.Errorf("incorrect deptype: %q", c.DepType)
	}
}

func TestExtractClauses_Grainsize(t *testing.T) {
	clauses, err := extractClauses("grainsize(5)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c, ok := clauses[0].(GrainsizeClause)
	if !ok {
		t.Fatalf("expected GrainsizeClause, got %T", clauses[0])
	}
	if c.Size != "5" {
		t.Errorf("incorrect size: %q", c.Size)
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

// =============================================================================
// Directive parsing - parseDirectiveText() in isolation
// =============================================================================

// --- Valid directives ---

func TestParseDirectiveText_Parallel(t *testing.T) {
	dir, err := parseDirectiveText("parallel shared(global) private(local)", 0, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	d, ok := dir.(ParallelDirective)
	if !ok {
		t.Fatalf("expected ParallelDirective, got %T", dir)
	}
	if len(d.Clauses) != 2 {
		t.Errorf("expected 2 clauses, got %d", len(d.Clauses))
	}
}

func TestParseDirectiveText_For(t *testing.T) {
	dir, err := parseDirectiveText("for private(i)", 0, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	d, ok := dir.(ForDirective)
	if !ok {
		t.Fatalf("expected ForDirective, got %T", dir)
	}
	if len(d.Clauses) != 1 {
		t.Errorf("expected 1 clause, got %d", len(d.Clauses))
	}
}

func TestParseDirectiveText_ParallelFor(t *testing.T) {
	dir, err := parseDirectiveText("parallel for schedule(static, 10) private(x)", 0, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	d, ok := dir.(ParallelForDirective)
	if !ok {
		t.Fatalf("expected ParallelForDirective, got %T", dir)
	}
	if len(d.Clauses) != 2 {
		t.Errorf("expected 2 clauses, got %d", len(d.Clauses))
	}
}

func TestParseDirectiveText_Single(t *testing.T) {
	dir, err := parseDirectiveText("single firstprivate(x)", 0, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := dir.(SingleDirective); !ok {
		t.Fatalf("expected SingleDirective, got %T", dir)
	}
}

func TestParseDirectiveText_Barrier(t *testing.T) {
	dir, err := parseDirectiveText("barrier", 0, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := dir.(BarrierDirective); !ok {
		t.Fatalf("expected BarrierDirective, got %T", dir)
	}
}

func TestParseDirectiveText_Taskwait(t *testing.T) {
	dir, err := parseDirectiveText("taskwait", 0, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := dir.(TaskwaitDirective); !ok {
		t.Fatalf("expected TaskwaitDirective, got %T", dir)
	}
}

func TestParseDirectiveText_Taskgroup(t *testing.T) {
	dir, err := parseDirectiveText("taskgroup", 0, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := dir.(TaskgroupDirective); !ok {
		t.Fatalf("expected TaskgroupDirective, got %T", dir)
	}
}

func TestParseDirectiveText_CriticalNamed(t *testing.T) {
	dir, err := parseDirectiveText("critical(mylock)", 0, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	d, ok := dir.(CriticalDirective)
	if !ok {
		t.Fatalf("expected CriticalDirective, got %T", dir)
	}
	if d.Name != "mylock" {
		t.Errorf("incorrect name: %q", d.Name)
	}
}

func TestParseDirectiveText_CriticalAnonymous(t *testing.T) {
	dir, err := parseDirectiveText("critical", 0, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	d, ok := dir.(CriticalDirective)
	if !ok {
		t.Fatalf("expected CriticalDirective, got %T", dir)
	}
	if d.Name != "" {
		t.Errorf("anonymous critical should have empty name, got %q", d.Name)
	}
}

func TestParseDirectiveText_AtomicUpdate(t *testing.T) {
	dir, err := parseDirectiveText("atomic update", 0, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	d, ok := dir.(AtomicDirective)
	if !ok {
		t.Fatalf("expected AtomicDirective, got %T", dir)
	}
	if d.Mode != "update" {
		t.Errorf("incorrect mode: %q", d.Mode)
	}
}

func TestParseDirectiveText_AtomicRead(t *testing.T) {
	dir, err := parseDirectiveText("atomic read", 0, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	d, ok := dir.(AtomicDirective)
	if !ok {
		t.Fatalf("expected AtomicDirective, got %T", dir)
	}
	if d.Mode != "read" {
		t.Errorf("incorrect mode: %q", d.Mode)
	}
}

func TestParseDirectiveText_AtomicWrite(t *testing.T) {
	dir, err := parseDirectiveText("atomic write", 0, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	d, ok := dir.(AtomicDirective)
	if !ok {
		t.Fatalf("expected AtomicDirective, got %T", dir)
	}
	if d.Mode != "write" {
		t.Errorf("incorrect mode: %q", d.Mode)
	}
}

func TestParseDirectiveText_AtomicDefaultMode(t *testing.T) {
	dir, err := parseDirectiveText("atomic", 0, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	d, ok := dir.(AtomicDirective)
	if !ok {
		t.Fatalf("expected AtomicDirective, got %T", dir)
	}
	if d.Mode != "" {
		t.Errorf("expected empty mode for default atomic, got %q", d.Mode)
	}
}

func TestParseDirectiveText_Task(t *testing.T) {
	dir, err := parseDirectiveText("task depend(out:buff) private(temp)", 0, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	d, ok := dir.(TaskDirective)
	if !ok {
		t.Fatalf("expected TaskDirective, got %T", dir)
	}
	if len(d.Clauses) != 2 {
		t.Errorf("expected 2 clauses, got %d", len(d.Clauses))
	}
}

func TestParseDirectiveText_Taskloop(t *testing.T) {
	dir, err := parseDirectiveText("taskloop grainsize(10)", 0, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := dir.(TaskloopDirective); !ok {
		t.Fatalf("expected TaskloopDirective, got %T", dir)
	}
}

// --- Generic error cases ---

func TestParseDirectiveText_Empty(t *testing.T) {
	_, err := parseDirectiveText("", 0, 1)
	if err == nil {
		t.Fatal("expected error for empty directive text")
	}
}

func TestParseDirectiveText_Unknown(t *testing.T) {
	_, err := parseDirectiveText("unknowndirective", 0, 1)
	if err == nil {
		t.Fatal("expected error for unknown directive")
	}
}

func TestParseDirectiveText_AtomicInvalidMode(t *testing.T) {
	_, err := parseDirectiveText("atomic invalidmode", 0, 1)
	if err == nil {
		t.Fatal("expected error for invalid atomic mode")
	}
}

func TestParseDirectiveText_CriticalMalformedName(t *testing.T) {
	_, err := parseDirectiveText("critical noparens", 0, 1)
	if err == nil {
		t.Fatal("expected error for critical name without parentheses")
	}
}

func TestParseDirectiveText_CriticalEmptyName(t *testing.T) {
	_, err := parseDirectiveText("critical()", 0, 1)
	if err == nil {
		t.Fatal("expected error for critical with empty name")
	}
}

// --- Rejection of clauses on directives that accept none ---

func TestParseDirectiveText_BarrierRejectsClause(t *testing.T) {
	_, err := parseDirectiveText("barrier private(x)", 0, 1)
	if err == nil {
		t.Fatal("expected error: barrier accepts no clauses")
	}
}

func TestParseDirectiveText_MasterRejectsClause(t *testing.T) {
	_, err := parseDirectiveText("master shared(y)", 0, 1)
	if err == nil {
		t.Fatal("expected error: master accepts no clauses")
	}
}

func TestParseDirectiveText_SectionRejectsClause(t *testing.T) {
	_, err := parseDirectiveText("section private(x)", 0, 1)
	if err == nil {
		t.Fatal("expected error: section accepts no clauses")
	}
}

func TestParseDirectiveText_TaskwaitRejectsClause(t *testing.T) {
	_, err := parseDirectiveText("taskwait private(x)", 0, 1)
	if err == nil {
		t.Fatal("expected error: taskwait accepts no clauses")
	}
}

func TestParseDirectiveText_TaskgroupRejectsClause(t *testing.T) {
	_, err := parseDirectiveText("taskgroup private(x)", 0, 1)
	if err == nil {
		t.Fatal("expected error: taskgroup accepts no clauses")
	}
}

// --- Rejection of clauses not allowed for the directive ---

func TestParseDirectiveText_ParallelRejectsDepend(t *testing.T) {
	_, err := parseDirectiveText("parallel depend(in:x)", 0, 1)
	if err == nil {
		t.Fatal("expected error: parallel does not accept depend")
	}
}

func TestParseDirectiveText_ParallelForRejectsDepend(t *testing.T) {
	_, err := parseDirectiveText("parallel for depend(in:x)", 0, 1)
	if err == nil {
		t.Fatal("expected error: parallel for does not accept depend")
	}
}

func TestParseDirectiveText_SectionsRejectsDepend(t *testing.T) {
	_, err := parseDirectiveText("sections depend(in:x)", 0, 1)
	if err == nil {
		t.Fatal("expected error: sections does not accept depend")
	}
}

func TestParseDirectiveText_ForRejectsReduction(t *testing.T) {
	_, err := parseDirectiveText("for reduction(+:x)", 0, 1)
	if err == nil {
		t.Fatal("expected error: for does not accept reduction")
	}
}

func TestParseDirectiveText_SingleRejectsShared(t *testing.T) {
	_, err := parseDirectiveText("single shared(x)", 0, 1)
	if err == nil {
		t.Fatal("expected error: single does not accept shared")
	}
}

func TestParseDirectiveText_TaskRejectsSchedule(t *testing.T) {
	_, err := parseDirectiveText("task schedule(static)", 0, 1)
	if err == nil {
		t.Fatal("expected error: task does not accept schedule")
	}
}

func TestParseDirectiveText_TaskloopRejectsShared(t *testing.T) {
	_, err := parseDirectiveText("taskloop shared(x)", 0, 1)
	if err == nil {
		t.Fatal("expected error: taskloop does not accept shared")
	}
}

// --- Rejection of unknown clauses for each directive that accepts clauses ---

func TestParseDirectiveText_ParallelInvalidClause(t *testing.T) {
	_, err := parseDirectiveText("parallel invalidclause(x)", 0, 1)
	if err == nil {
		t.Fatal("expected error: invalid clause in parallel")
	}
}

func TestParseDirectiveText_ForInvalidClause(t *testing.T) {
	_, err := parseDirectiveText("for invalidclause(x)", 0, 1)
	if err == nil {
		t.Fatal("expected error: invalid clause in for")
	}
}

func TestParseDirectiveText_ParallelForInvalidClause(t *testing.T) {
	_, err := parseDirectiveText("parallel for invalidclause(x)", 0, 1)
	if err == nil {
		t.Fatal("expected error: invalid clause in parallel for")
	}
}

func TestParseDirectiveText_SectionsInvalidClause(t *testing.T) {
	_, err := parseDirectiveText("sections invalidclause(x)", 0, 1)
	if err == nil {
		t.Fatal("expected error: invalid clause in sections")
	}
}

func TestParseDirectiveText_SingleInvalidClause(t *testing.T) {
	_, err := parseDirectiveText("single invalidclause(x)", 0, 1)
	if err == nil {
		t.Fatal("expected error: invalid clause in single")
	}
}

func TestParseDirectiveText_TaskInvalidClause(t *testing.T) {
	_, err := parseDirectiveText("task invalidclause(x)", 0, 1)
	if err == nil {
		t.Fatal("expected error: invalid clause in task")
	}
}

func TestParseDirectiveText_TaskloopInvalidClause(t *testing.T) {
	_, err := parseDirectiveText("taskloop invalidclause(x)", 0, 1)
	if err == nil {
		t.Fatal("expected error: invalid clause in taskloop")
	}
}

// =============================================================================
// Full integration - Parse() with real Go source code
// =============================================================================

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
	if _, ok := result.Nodes[0].Directive.(ParallelDirective); !ok {
		t.Errorf("expected ParallelDirective, got %T", result.Nodes[0].Directive)
	}
}

func TestParse_For(t *testing.T) {
	src := `package main

func main() {
	//gompher parallel
	{
		//gompher for
		for i := 0; i < 100; i++ {
		}
	}
}`
	result, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(result.Nodes))
	}
	if _, ok := result.Nodes[0].Directive.(ParallelDirective); !ok {
		t.Errorf("node 0: expected ParallelDirective, got %T", result.Nodes[0].Directive)
	}
	if _, ok := result.Nodes[1].Directive.(ForDirective); !ok {
		t.Errorf("node 1: expected ForDirective, got %T", result.Nodes[1].Directive)
	}
}

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
	if _, ok := result.Nodes[0].Directive.(ParallelForDirective); !ok {
		t.Errorf("expected ParallelForDirective, got %T", result.Nodes[0].Directive)
	}
}

func TestParse_Sections(t *testing.T) {
	src := `package main

func main() {
	//gompher parallel
	{
		//gompher sections
		{
			//gompher section
			{
				decodificarVideo()
			}

			//gompher section
			{
				decodificarAudio()
			}
		}
	}
}`
	result, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Nodes) != 4 {
		t.Fatalf("expected 4 nodes, got %d", len(result.Nodes))
	}
	if _, ok := result.Nodes[0].Directive.(ParallelDirective); !ok {
		t.Errorf("node 0: expected ParallelDirective, got %T", result.Nodes[0].Directive)
	}
	if _, ok := result.Nodes[1].Directive.(SectionsDirective); !ok {
		t.Errorf("node 1: expected SectionsDirective, got %T", result.Nodes[1].Directive)
	}
	if _, ok := result.Nodes[2].Directive.(SectionDirective); !ok {
		t.Errorf("node 2: expected SectionDirective, got %T", result.Nodes[2].Directive)
	}
	if _, ok := result.Nodes[3].Directive.(SectionDirective); !ok {
		t.Errorf("node 3: expected SectionDirective, got %T", result.Nodes[3].Directive)
	}
}

func TestParse_Single(t *testing.T) {
	src := `package main

func main() {
	//gompher parallel
	{
		//gompher single
		{
			log.Println("once")
		}
	}
}`
	result, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(result.Nodes))
	}
	if _, ok := result.Nodes[1].Directive.(SingleDirective); !ok {
		t.Errorf("node 1: expected SingleDirective, got %T", result.Nodes[1].Directive)
	}
}

func TestParse_Master(t *testing.T) {
	src := `package main

func main() {
	//gompher parallel
	{
		//gompher master
		{
			fmt.Println("master only")
		}
	}
}`
	result, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(result.Nodes))
	}
	if _, ok := result.Nodes[1].Directive.(MasterDirective); !ok {
		t.Errorf("node 1: expected MasterDirective, got %T", result.Nodes[1].Directive)
	}
}

func TestParse_Critical(t *testing.T) {
	src := `package main

func main() {
	//gompher parallel
	{
		//gompher critical
		{
			contador++
		}
	}
}`
	result, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(result.Nodes))
	}
	if _, ok := result.Nodes[1].Directive.(CriticalDirective); !ok {
		t.Errorf("node 1: expected CriticalDirective, got %T", result.Nodes[1].Directive)
	}
}

func TestParse_Atomic(t *testing.T) {
	src := `package main

func main() {
	//gompher parallel
	{
		//gompher atomic update
		contador++
	}
}`
	result, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(result.Nodes))
	}
	d, ok := result.Nodes[1].Directive.(AtomicDirective)
	if !ok {
		t.Fatalf("expected AtomicDirective, got %T", result.Nodes[1].Directive)
	}
	if d.Mode != "update" {
		t.Errorf("incorrect mode: %q", d.Mode)
	}
}

func TestParse_Tasks(t *testing.T) {
	src := `package main

func main() {
	//gompher task depend(in:x)
	{
		procesar(x)
	}

	//gompher taskwait
}`
	result, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(result.Nodes))
	}

	if _, ok := result.Nodes[0].Directive.(TaskDirective); !ok {
		t.Errorf("node 0: expected TaskDirective, got %T", result.Nodes[0].Directive)
	}

	if _, ok := result.Nodes[1].Directive.(TaskwaitDirective); !ok {
		t.Errorf("node 1: expected TaskwaitDirective, got %T", result.Nodes[1].Directive)
	}
}

func TestParse_Taskgroup(t *testing.T) {
	src := `package main

func main() {
	//gompher taskgroup
	{
		crearArbolRecursivo()
	}
}`
	result, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(result.Nodes))
	}
	if _, ok := result.Nodes[0].Directive.(TaskgroupDirective); !ok {
		t.Errorf("expected TaskgroupDirective, got %T", result.Nodes[0].Directive)
	}
}

func TestParse_Taskloop(t *testing.T) {
	src := `package main

func main() {
	//gompher taskloop grainsize(5)
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
	d, ok := result.Nodes[0].Directive.(TaskloopDirective)
	if !ok {
		t.Fatalf("expected TaskloopDirective, got %T", result.Nodes[0].Directive)
	}
	if len(d.Clauses) != 1 {
		t.Errorf("expected 1 clause, got %d", len(d.Clauses))
	}
}

func TestParse_MultipleDirectives(t *testing.T) {
	src := `package main

func main() {
	//gompher parallel for private(x)
	for i := 0; i < 100; i++ {
	}

	//gompher barrier

	//gompher parallel for reduction(+:suma)
	for i := 0; i < 10; i++ {
	}
}`
	result, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(result.Nodes))
	}
	if _, ok := result.Nodes[0].Directive.(ParallelForDirective); !ok {
		t.Errorf("node 0: expected ParallelForDirective, got %T", result.Nodes[0].Directive)
	}
	if _, ok := result.Nodes[1].Directive.(BarrierDirective); !ok {
		t.Errorf("node 1: expected BarrierDirective, got %T", result.Nodes[1].Directive)
	}
	if _, ok := result.Nodes[2].Directive.(ParallelForDirective); !ok {
		t.Errorf("node 2: expected ParallelForDirective, got %T", result.Nodes[2].Directive)
	}
}

func TestParse_NonGompherCommentIgnored(t *testing.T) {
	src := `package main

func main() {
	// this is a regular comment
	//gompher parallel
	{
		doWork()
	}
}`
	result, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(result.Nodes))
	}
	if _, ok := result.Nodes[0].Directive.(ParallelDirective); !ok {
		t.Errorf("expected ParallelDirective, got %T", result.Nodes[0].Directive)
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

func TestParse_InvalidGompherDirective(t *testing.T) {
	src := `package main

func main() {
	//gompher unknowndirective
	{}
}`
	_, err := Parse(src)
	if err == nil {
		t.Fatal("expected error for invalid gompher directive")
	}
}

// =============================================================================
// Semantic validations
// =============================================================================

// --- Node type validation ---

func TestParse_ForOnNonForStmt(t *testing.T) {
	src := `package main

func main() {
	//gompher for
	{
		x := 1
		_ = x
	}
}`
	_, err := Parse(src)
	if err == nil {
		t.Fatal("expected error: //gompher for must target a for loop")
	}
}

func TestParse_ParallelForOnNonForStmt(t *testing.T) {
	src := `package main

func main() {
	//gompher parallel for
	{
		x := 1
		_ = x
	}
}`
	_, err := Parse(src)
	if err == nil {
		t.Fatal("expected error: //gompher parallel for must target a for loop")
	}
}

func TestParse_TaskloopOnNonForStmt(t *testing.T) {
	src := `package main

func main() {
	//gompher taskloop
	{
		x := 1
		_ = x
	}
}`
	_, err := Parse(src)
	if err == nil {
		t.Fatal("expected error: //gompher taskloop must target a for loop")
	}
}

func TestParse_ParallelOnNonBlockStmt(t *testing.T) {
	src := `package main

func main() {
	//gompher parallel
	for i := 0; i < 10; i++ {
	}
}`
	_, err := Parse(src)
	if err == nil {
		t.Fatal("expected error: //gompher parallel must target a block statement")
	}
}

func TestParse_AtomicOnBlockStmt(t *testing.T) {
	src := `package main

func main() {
	//gompher atomic
	{
		x := 1
		_ = x
	}
}`
	_, err := Parse(src)
	if err == nil {
		t.Fatal("expected error: //gompher atomic must target an expression or assignment")
	}
}

func TestParse_AtomicOnAssignStmt(t *testing.T) {
	src := `package main

var contador int

func main() {
	//gompher atomic update
	contador = contador + 1
}`
	result, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.Nodes[0].Directive.(AtomicDirective); !ok {
		t.Errorf("expected AtomicDirective, got %T", result.Nodes[0].Directive)
	}
}

// --- Adjacency validation (blank line between directive and target) ---

func TestParse_BlankLineBetweenDirectiveAndBlock(t *testing.T) {
	src := `package main

func main() {
	//gompher parallel

	{
		x := 1
		_ = x
	}
}`
	_, err := Parse(src)
	if err == nil {
		t.Fatal("expected error: blank line between directive and block")
	}
}

// --- Hierarchical context validation (section must be inside sections) ---

func TestParse_SectionOutsideSections(t *testing.T) {
	src := `package main

func main() {
	//gompher section
	{
		doWork()
	}
}`
	_, err := Parse(src)
	if err == nil {
		t.Fatal("expected error: //gompher section must appear inside //gompher sections")
	}
}

// --- Clause validation (empty argument lists) ---

func TestExtractClauses_EmptyPrivate(t *testing.T) {
	_, err := extractClauses("private()")
	if err == nil {
		t.Fatal("expected error for private()")
	}
	if !strings.Contains(err.Error(), "at least one variable") {
		t.Errorf("expected message about empty variable list, got: %v", err)
	}
}

func TestExtractClauses_EmptyShared(t *testing.T) {
	_, err := extractClauses("shared(  )")
	if err == nil {
		t.Fatal("expected error for shared() with whitespace")
	}
}

// =============================================================================
// Interface contracts - verify every concrete type implements its interface
// =============================================================================

func TestDirectiveKind_AllTypes(t *testing.T) {
	cases := []struct {
		dir      Directive
		expected DirectiveKind
	}{
		{ParallelDirective{}, DirParallel},
		{ForDirective{}, DirFor},
		{ParallelForDirective{}, DirParallelFor},
		{SectionsDirective{}, DirSections},
		{SectionDirective{}, DirSection},
		{SingleDirective{}, DirSingle},
		{MasterDirective{}, DirMaster},
		{CriticalDirective{}, DirCritical},
		{BarrierDirective{}, DirBarrier},
		{AtomicDirective{}, DirAtomic},
		{TaskDirective{}, DirTask},
		{TaskwaitDirective{}, DirTaskwait},
		{TaskgroupDirective{}, DirTaskgroup},
		{TaskloopDirective{}, DirTaskloop},
	}
	for _, tc := range cases {
		if got := tc.dir.directiveKind(); got != tc.expected {
			t.Errorf("%T: expected kind %q, got %q", tc.dir, tc.expected, got)
		}
	}
}

func TestDirectiveLine_AllTypes(t *testing.T) {
	cases := []struct {
		dir      Directive
		expected int
	}{
		{ParallelDirective{pos: pos{Line: 1}}, 1},
		{ForDirective{pos: pos{Line: 2}}, 2},
		{ParallelForDirective{pos: pos{Line: 3}}, 3},
		{SectionsDirective{pos: pos{Line: 4}}, 4},
		{SectionDirective{pos: pos{Line: 5}}, 5},
		{SingleDirective{pos: pos{Line: 6}}, 6},
		{MasterDirective{pos: pos{Line: 7}}, 7},
		{CriticalDirective{pos: pos{Line: 8}}, 8},
		{BarrierDirective{pos: pos{Line: 9}}, 9},
		{AtomicDirective{pos: pos{Line: 10}}, 10},
		{TaskDirective{pos: pos{Line: 11}}, 11},
		{TaskwaitDirective{pos: pos{Line: 12}}, 12},
		{TaskgroupDirective{pos: pos{Line: 13}}, 13},
		{TaskloopDirective{pos: pos{Line: 14}}, 14},
	}
	for _, tc := range cases {
		if got := tc.dir.line(); got != tc.expected {
			t.Errorf("%T: expected line %d, got %d", tc.dir, tc.expected, got)
		}
	}
}

func TestClauseKind_AllTypes(t *testing.T) {
	cases := []struct {
		clause   Clause
		expected ClauseKind
	}{
		{PrivateClause{}, ClausePrivate},
		{FirstPrivateClause{}, ClauseFirstPrivate},
		{LastPrivateClause{}, ClauseLastPrivate},
		{SharedClause{}, ClauseShared},
		{ReductionClause{}, ClauseReduction},
		{ScheduleClause{}, ClauseSchedule},
		{DependClause{}, ClauseDepend},
		{GrainsizeClause{}, ClauseGrainsize},
	}
	for _, tc := range cases {
		if got := tc.clause.clauseKind(); got != tc.expected {
			t.Errorf("%T: expected kind %q, got %q", tc.clause, tc.expected, got)
		}
	}
}

// =============================================================================
// Internal helpers - direct calls to unexported functions
// =============================================================================

func TestValidateClauses_KindNotInMapWithClauses(t *testing.T) {
	err := validateClauses(DirBarrier, []Clause{PrivateClause{Vars: []string{"x"}}})
	if err == nil {
		t.Fatal("expected error: barrier not in validClauses map")
	}
}

func TestValidateClauses_KindNotInMapNoClauses(t *testing.T) {
	err := validateClauses(DirBarrier, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMakeVarListClause_UnknownKind(t *testing.T) {
	_, err := makeVarListClause("impossible", nil)
	if err == nil {
		t.Fatal("expected error for unknown clause kind")
	}
}

func TestBuildDirective_UnknownKind(t *testing.T) {
	_, err := buildDirective(DirectiveKind("bogusss"), "", pos{})
	if err == nil {
		t.Fatal("expected error for unknown kind")
	}
}
