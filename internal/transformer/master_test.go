package transformer

import (
	"go/ast"
	"strings"
	"testing"

	"github.com/gomphermp/gomphermp/internal/parser"
)

// TestTransform_Master_BasicRewrite verifies the canonical master
// transformation: //gompher master { body } becomes
// runtime.Master(threadID, func() { body }). The threadID identifier is
// emitted unqualified so it resolves to whichever int parameter is in scope
// (typically the one introduced by the surrounding runtime.Parallel call).
func TestTransform_Master_BasicRewrite(t *testing.T) {
	src := `package main

func worker(threadID int) {
	//gompher master
	{
		log_thread_zero()
	}
}

func log_thread_zero() {}
`
	got := runTransform(t, src)

	if !strings.Contains(got, `runtime.Master(threadID, func() {`) {
		t.Errorf("expected runtime.Master(threadID, func() {...}) in output, got:\n%s", got)
	}
	if !strings.Contains(got, "log_thread_zero()") {
		t.Errorf("expected original body call preserved, got:\n%s", got)
	}
}

// TestTransform_Master_AddsRuntimeImport verifies that the runtime import
// is injected when a master directive is rewritten. Master shares the
// import-injection logic with Critical and Single via Transform's
// emittedRuntime flag; this test confirms it triggers via the master path.
func TestTransform_Master_AddsRuntimeImport(t *testing.T) {
	src := `package main

func worker(threadID int) {
	//gompher master
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

// TestTransform_Master_PreservesMultiStmtBody verifies that the body is
// moved verbatim into the closure. Master's body, like Single's and
// Critical's, is treated as an opaque sequence of statements in this phase.
func TestTransform_Master_PreservesMultiStmtBody(t *testing.T) {
	src := `package main

import "fmt"

func worker(threadID int) {
	//gompher master
	{
		header := "==="
		count := 42
		fmt.Println(header, count)
	}
}
`
	got := runTransform(t, src)

	for _, want := range []string{`header := "==="`, "count := 42", "fmt.Println(header, count)"} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q preserved, got:\n%s", want, got)
		}
	}
}

// TestTransform_Master_PassesThreadIDByName verifies that the transformer
// always emits the identifier threadID as the first argument, regardless of
// any reformatting the body may undergo. The constant threadIDParamName in
// master.go is the single source of truth; this test asserts it is the name
// actually emitted in the output.
func TestTransform_Master_PassesThreadIDByName(t *testing.T) {
	src := `package main

func worker(threadID int) {
	//gompher master
	{
		x := 1
		_ = x
	}
}
`
	got := runTransform(t, src)

	if !strings.Contains(got, `runtime.Master(threadID,`) {
		t.Errorf("expected first arg to runtime.Master to be the identifier %q, got:\n%s", threadIDParamName, got)
	}
}

// TestTransformMaster_WrongNodeType verifies the defensive error path
// inherited from transformBlockDirective: passing a non-BlockStmt Node
// surfaces a descriptive error instead of panicking. Mirror of the same
// check in critical and single.
func TestTransformMaster_WrongNodeType(t *testing.T) {
	parsed, err := parser.Parse("package main\n")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	bogus := parser.MasterDirective{
		Node: &ast.ExprStmt{},
	}

	err = transformMaster(parsed, bogus)
	if err == nil {
		t.Fatal("expected error for non-BlockStmt Node")
	}
	if !strings.Contains(err.Error(), "expected *ast.BlockStmt") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestTransform_PropagatesMasterError verifies that when transformMaster
// fails, Transform propagates the error and returns nil instead of a
// partially transformed file. Mirror of the equivalent contract test for
// Critical and Single, ensuring the error-handling branch in Transform's
// dispatch is exercised for MasterDirective.
func TestTransform_PropagatesMasterError(t *testing.T) {
	parsed, err := parser.Parse("package main\n")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	parsed.Nodes = append(parsed.Nodes, parser.AnnotatedNode{
		Directive: parser.MasterDirective{
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
