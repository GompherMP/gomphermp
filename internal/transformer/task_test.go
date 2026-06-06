package transformer

import (
	"go/ast"
	"strings"
	"testing"

	"github.com/gomphermp/gomphermp/internal/parser"
)

// TestTransform_Task_BasicRewrite verifies the canonical task transformation:
// //gompher task { body } becomes runtime.Task(func() { body }).
func TestTransform_Task_BasicRewrite(t *testing.T) {
	src := `package main

func main() {
	//gompher task
	{
		compute()
	}
}

func compute() {}
`
	got := runTransform(t, src)

	if !strings.Contains(got, `runtime.Task(func() {`) {
		t.Errorf("expected runtime.Task(func() {...}) in output, got:\n%s", got)
	}
	if !strings.Contains(got, "compute()") {
		t.Errorf("expected body call preserved, got:\n%s", got)
	}
}

// TestTransform_Task_AddsRuntimeImport verifies that the runtime import is
// injected when a task directive appears in a file with no prior imports.
func TestTransform_Task_AddsRuntimeImport(t *testing.T) {
	src := `package main

func main() {
	//gompher task
	{
		work()
	}
}

func work() {}
`
	got := runTransform(t, src)

	if !strings.Contains(got, `"github.com/gomphermp/gomphermp/pkg/runtime"`) {
		t.Errorf("expected runtime import in output, got:\n%s", got)
	}
}

// TestTransform_Task_PreservesMultiStmtBody verifies that a task body with
// several statements is moved verbatim into the closure.
func TestTransform_Task_PreservesMultiStmtBody(t *testing.T) {
	src := `package main

import "fmt"

func main() {
	//gompher task
	{
		a := 1
		b := 2
		fmt.Println(a + b)
	}
}
`
	got := runTransform(t, src)

	for _, want := range []string{"a := 1", "b := 2", "fmt.Println(a + b)"} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q preserved, got:\n%s", want, got)
		}
	}
}

// TestTransform_Task_DependIn verifies that a task with depend(in:x) becomes
// runtime.TaskWithDepend(func(){...}, []uintptr{uintptr(unsafe.Pointer(&x))}, nil, nil).
func TestTransform_Task_DependIn(t *testing.T) {
	src := `package main

func main() {
	x := 0
	//gompher task depend(in:x)
	{
		_ = x
	}
	_ = x
}
`
	got := runTransform(t, src)

	if !strings.Contains(got, "runtime.TaskWithDepend(") {
		t.Errorf("expected runtime.TaskWithDepend in output, got:\n%s", got)
	}
	if !strings.Contains(got, "unsafe.Pointer(&x)") {
		t.Errorf("expected unsafe.Pointer(&x) in output, got:\n%s", got)
	}
	if !strings.Contains(got, `"unsafe"`) {
		t.Errorf("expected unsafe import in output, got:\n%s", got)
	}
}

// TestTransform_Task_DependOut verifies that depend(out:y) builds the outs
// slice and passes nil for ins and inouts.
func TestTransform_Task_DependOut(t *testing.T) {
	src := `package main

func main() {
	y := 0
	//gompher task depend(out:y)
	{
		y = 1
	}
	_ = y
}
`
	got := runTransform(t, src)

	if !strings.Contains(got, "runtime.TaskWithDepend(") {
		t.Errorf("expected runtime.TaskWithDepend in output, got:\n%s", got)
	}
	if !strings.Contains(got, "unsafe.Pointer(&y)") {
		t.Errorf("expected unsafe.Pointer(&y) in output, got:\n%s", got)
	}
}

// TestTransform_Task_DependAllTypes verifies that depend(in:x) depend(out:y)
// depend(inout:z) populates all three slices in the correct order.
func TestTransform_Task_DependAllTypes(t *testing.T) {
	src := `package main

func main() {
	x, y, z := 0, 0, 0
	//gompher task depend(in:x) depend(out:y) depend(inout:z)
	{
		y = x + z
		z++
	}
	_, _, _ = x, y, z
}
`
	got := runTransform(t, src)

	if !strings.Contains(got, "runtime.TaskWithDepend(") {
		t.Errorf("expected runtime.TaskWithDepend in output, got:\n%s", got)
	}
	for _, v := range []string{"&x", "&y", "&z"} {
		if !strings.Contains(got, v) {
			t.Errorf("expected %s in output, got:\n%s", v, got)
		}
	}
}

// TestTransform_Taskwait_Basic verifies that //gompher taskwait is rewritten
// to runtime.Taskwait() inserted at the correct position.
func TestTransform_Taskwait_Basic(t *testing.T) {
	src := `package main

func main() {
	//gompher task
	{
		compute()
	}
	//gompher taskwait
	process()
}

func compute() {}
func process() {}
`
	got := runTransform(t, src)

	if !strings.Contains(got, "runtime.Taskwait()") {
		t.Errorf("expected runtime.Taskwait() in output, got:\n%s", got)
	}
	// Taskwait must appear before process()
	twIdx := strings.Index(got, "runtime.Taskwait()")
	procIdx := strings.Index(got, "process()")
	if twIdx == -1 || procIdx == -1 || twIdx > procIdx {
		t.Errorf("expected Taskwait() before process(), got:\n%s", got)
	}
}

// TestTransform_Taskwait_AtEndOfBlock verifies that taskwait at the end of a
// block (no following statement) is appended correctly.
func TestTransform_Taskwait_AtEndOfBlock(t *testing.T) {
	src := `package main

func main() {
	//gompher task
	{
		compute()
	}
	//gompher taskwait
}

func compute() {}
`
	got := runTransform(t, src)

	if !strings.Contains(got, "runtime.Taskwait()") {
		t.Errorf("expected runtime.Taskwait() in output, got:\n%s", got)
	}
}

// TestTransform_Taskgroup_BasicRewrite verifies that //gompher taskgroup
// { body } becomes runtime.Taskgroup(func() { body }).
func TestTransform_Taskgroup_BasicRewrite(t *testing.T) {
	src := `package main

func main() {
	//gompher taskgroup
	{
		//gompher task
		{
			work()
		}
	}
}

func work() {}
`
	got := runTransform(t, src)

	if !strings.Contains(got, "runtime.Taskgroup(func() {") {
		t.Errorf("expected runtime.Taskgroup(func() {...}) in output, got:\n%s", got)
	}
}

// TestTransform_Taskgroup_AddsRuntimeImport verifies that the runtime import
// is injected when a taskgroup directive appears in a file with no prior imports.
func TestTransform_Taskgroup_AddsRuntimeImport(t *testing.T) {
	src := `package main

func main() {
	//gompher taskgroup
	{
		work()
	}
}

func work() {}
`
	got := runTransform(t, src)

	if !strings.Contains(got, `"github.com/gomphermp/gomphermp/pkg/runtime"`) {
		t.Errorf("expected runtime import in output, got:\n%s", got)
	}
}

// TestTransformTask_WrongNodeType verifies the defensive error path when
// the directive's Node is not a *ast.BlockStmt.
func TestTransformTask_WrongNodeType(t *testing.T) {
	parsed, err := parser.Parse("package main\n")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	// depend path requires BlockStmt — use a depend clause to force that branch
	bogus := parser.TaskDirective{
		Clauses: []parser.Clause{
			parser.DependClause{DepType: "in", Vars: []string{"x"}},
		},
		Node: &ast.ExprStmt{},
	}

	err = transformTask(parsed, bogus)
	if err == nil {
		t.Fatal("expected error for non-BlockStmt Node")
	}
	if !strings.Contains(err.Error(), "expected *ast.BlockStmt") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestTransformTask_BodyNotInAST verifies that when the task's BlockStmt is
// not reachable from the file's AST, the transformer reports the inconsistency.
func TestTransformTask_BodyNotInAST(t *testing.T) {
	parsed, err := parser.Parse("package main\n")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	detached := &ast.BlockStmt{}
	bogus := parser.TaskDirective{
		Clauses: []parser.Clause{
			parser.DependClause{DepType: "out", Vars: []string{"y"}},
		},
		Node: detached,
	}

	err = transformTask(parsed, bogus)
	if err == nil {
		t.Fatal("expected error when body block is detached from AST")
	}
	if !strings.Contains(err.Error(), "body block not found") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestTransformTaskgroup_WrongNodeType verifies the defensive error path for
// taskgroup when Node is not a *ast.BlockStmt.
func TestTransformTaskgroup_WrongNodeType(t *testing.T) {
	parsed, err := parser.Parse("package main\n")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	bogus := parser.TaskgroupDirective{
		Node: &ast.ExprStmt{},
	}

	err = transformTaskgroup(parsed, bogus)
	if err == nil {
		t.Fatal("expected error for non-BlockStmt Node")
	}
	if !strings.Contains(err.Error(), "expected *ast.BlockStmt") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestTransformTask_NoDependWrongNodeType verifies the defensive error path in
// the no-depend branch: transformBlockDirective must reject a non-BlockStmt Node
// even when no depend clauses are present.
func TestTransformTask_NoDependWrongNodeType(t *testing.T) {
	parsed, err := parser.Parse("package main\n")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	bogus := parser.TaskDirective{
		Node: &ast.ExprStmt{}, // no Clauses — routes through transformBlockDirective
	}

	err = transformTask(parsed, bogus)
	if err == nil {
		t.Fatal("expected error for non-BlockStmt Node in no-depend path")
	}
	if !strings.Contains(err.Error(), "expected *ast.BlockStmt") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestTransform_Task_DependArgumentOrder verifies that ins, outs, and inouts
// appear in that order as arguments to runtime.TaskWithDepend. The runtime
// relies on positional semantics — a swap would silently corrupt dependency
// tracking.
func TestTransform_Task_DependArgumentOrder(t *testing.T) {
	src := `package main

func main() {
	a, b, c := 0, 0, 0
	//gompher task depend(in:a) depend(out:b) depend(inout:c)
	{
		b = a + c
		c++
	}
	_, _, _ = a, b, c
}
`
	got := runTransform(t, src)

	aIdx := strings.Index(got, "&a")
	bIdx := strings.Index(got, "&b")
	cIdx := strings.Index(got, "&c")

	if aIdx == -1 || bIdx == -1 || cIdx == -1 {
		t.Fatalf("expected &a, &b, &c in output, got:\n%s", got)
	}
	if !(aIdx < bIdx && bIdx < cIdx) {
		t.Errorf("expected &a (in) < &b (out) < &c (inout) by position, got %d, %d, %d in:\n%s",
			aIdx, bIdx, cIdx, got)
	}
}

// TestTransform_Task_DependMultipleVarsInClause verifies that a single depend
// clause listing multiple variables (depend(in:x, y)) places all of them in
// the same ins slice.
func TestTransform_Task_DependMultipleVarsInClause(t *testing.T) {
	src := `package main

func main() {
	x, y := 0, 0
	//gompher task depend(in:x, y)
	{
		_ = x + y
	}
	_, _ = x, y
}
`
	got := runTransform(t, src)

	if !strings.Contains(got, "runtime.TaskWithDepend(") {
		t.Errorf("expected runtime.TaskWithDepend in output, got:\n%s", got)
	}
	if !strings.Contains(got, "unsafe.Pointer(&x)") {
		t.Errorf("expected &x in output, got:\n%s", got)
	}
	if !strings.Contains(got, "unsafe.Pointer(&y)") {
		t.Errorf("expected &y in output, got:\n%s", got)
	}
	// Both must be in the same slice, so &x must appear before the first nil
	xIdx := strings.Index(got, "&x")
	yIdx := strings.Index(got, "&y")
	firstNil := strings.Index(got, "nil")
	if firstNil != -1 && (xIdx > firstNil || yIdx > firstNil) {
		t.Errorf("expected &x and &y before first nil (same slice), got:\n%s", got)
	}
}

// TestTransform_Task_DependNilForEmptyGroups verifies that empty dependency
// groups emit nil rather than an empty []uintptr{} literal, keeping the
// generated code readable and matching the runtime's nil-tolerant API.
func TestTransform_Task_DependNilForEmptyGroups(t *testing.T) {
	src := `package main

func main() {
	x := 0
	//gompher task depend(in:x)
	{
		_ = x
	}
	_ = x
}
`
	got := runTransform(t, src)

	// outs and inouts should both be nil (2 nil arguments)
	nilCount := strings.Count(got, "nil")
	if nilCount < 2 {
		t.Errorf("expected at least 2 nil args for empty outs/inouts groups, got %d in:\n%s", nilCount, got)
	}
	// Must not emit empty slice literals
	if strings.Contains(got, "[]uintptr{}") {
		t.Errorf("expected nil not []uintptr{} for empty groups, got:\n%s", got)
	}
}

// TestTransform_Task_DependUnsafeImportIdempotent verifies that two task
// directives with depend clauses in the same file produce exactly one "unsafe"
// import, not two.
func TestTransform_Task_DependUnsafeImportIdempotent(t *testing.T) {
	src := `package main

func main() {
	x, y := 0, 0
	//gompher task depend(in:x)
	{
		_ = x
	}
	//gompher task depend(out:y)
	{
		y = 1
	}
	_, _ = x, y
}
`
	got := runTransform(t, src)

	count := strings.Count(got, `"unsafe"`)
	if count != 1 {
		t.Errorf("expected exactly 1 unsafe import, got %d in:\n%s", count, got)
	}
}

// TestTransform_Taskgroup_NestedTaskTransformed verifies that both the outer
// taskgroup and the inner task directive are fully rewritten when they are
// composed together.
func TestTransform_Taskgroup_NestedTaskTransformed(t *testing.T) {
	src := `package main

func main() {
	//gompher taskgroup
	{
		//gompher task
		{
			work()
		}
	}
}

func work() {}
`
	got := runTransform(t, src)

	if !strings.Contains(got, "runtime.Taskgroup(func() {") {
		t.Errorf("expected runtime.Taskgroup in output, got:\n%s", got)
	}
	if !strings.Contains(got, "runtime.Task(func() {") {
		t.Errorf("expected runtime.Task nested inside Taskgroup, got:\n%s", got)
	}
}

// TestTransform_Task_DirectiveCommentRemoved verifies that the //gompher task
// comment is stripped from the output after transformation.
func TestTransform_Task_DirectiveCommentRemoved(t *testing.T) {
	src := `package main

func main() {
	//gompher task
	{
		work()
	}
}

func work() {}
`
	got := runTransform(t, src)

	if strings.Contains(got, "//gompher") {
		t.Errorf("expected directive comment removed from output, got:\n%s", got)
	}
}

// TestTransform_Taskwait_DirectiveCommentRemoved verifies that the
// //gompher taskwait comment is stripped from the output after transformation.
func TestTransform_Taskwait_DirectiveCommentRemoved(t *testing.T) {
	src := `package main

func main() {
	//gompher task
	{
		compute()
	}
	//gompher taskwait
	process()
}

func compute() {}
func process() {}
`
	got := runTransform(t, src)

	if strings.Contains(got, "//gompher") {
		t.Errorf("expected directive comment removed from output, got:\n%s", got)
	}
}

// TestTransform_Taskgroup_DirectiveCommentRemoved verifies that the
// //gompher taskgroup comment is stripped from the output after transformation.
func TestTransform_Taskgroup_DirectiveCommentRemoved(t *testing.T) {
	src := `package main

func main() {
	//gompher taskgroup
	{
		work()
	}
}

func work() {}
`
	got := runTransform(t, src)

	if strings.Contains(got, "//gompher") {
		t.Errorf("expected directive comment removed from output, got:\n%s", got)
	}
}

// TestTransform_PropagatesTaskError verifies that Transform propagates task
// transformer errors and returns nil instead of a partially transformed file.
func TestTransform_PropagatesTaskError(t *testing.T) {
	parsed, err := parser.Parse("package main\n")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	parsed.Nodes = append(parsed.Nodes, parser.AnnotatedNode{
		Directive: parser.TaskDirective{
			Clauses: []parser.Clause{
				parser.DependClause{DepType: "in", Vars: []string{"x"}},
			},
			Node: &ast.BlockStmt{},
		},
	})

	got, err := Transform(parsed)
	if err == nil {
		t.Fatal("expected Transform to propagate the underlying error")
	}
	if got != nil {
		t.Errorf("expected nil ParseResult on error, got %v", got)
	}
}
