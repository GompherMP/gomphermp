package transformer

import (
	"strings"
	"testing"

	"github.com/gomphermp/gomphermp/internal/parser"
)

// TestTransform_Barrier_BasicInjection verifies that a //gompher barrier with
// no associated body is rewritten into a runtime.Barrier() call spliced into
// the enclosing block, between the statements that surround it.
func TestTransform_Barrier_BasicInjection(t *testing.T) {
	src := `package main

func worker() {
	phaseOne()
	//gompher barrier
	phaseTwo()
}

func phaseOne() {}
func phaseTwo() {}
`
	got := runTransform(t, src)

	if !strings.Contains(got, "runtime.Barrier()") {
		t.Errorf("expected runtime.Barrier() in output, got:\n%s", got)
	}
	// The barrier must land between the two phases.
	one := strings.Index(got, "phaseOne()")
	bar := strings.Index(got, "runtime.Barrier()")
	two := strings.Index(got, "phaseTwo()")
	if !(one < bar && bar < two) {
		t.Errorf("expected Barrier between phaseOne and phaseTwo; got one=%d bar=%d two=%d\n%s", one, bar, two, got)
	}
}

// TestTransform_Barrier_InsideParallel verifies the critical composition: a
// barrier nested inside a parallel region. After both directives are
// transformed, runtime.Barrier() must sit inside the runtime.Parallel closure
// (not as a sibling), proving the positional injection still locates the right
// block after the parallel rewrite reused the original body.
func TestTransform_Barrier_InsideParallel(t *testing.T) {
	src := `package main

func main() {
	//gompher parallel
	{
		initPhase()
		//gompher barrier
		computePhase()
	}
}

func initPhase() {}
func computePhase() {}
`
	got := runTransform(t, src)

	if !strings.Contains(got, `runtime.Parallel(func(threadID int) {`) {
		t.Errorf("expected Parallel closure, got:\n%s", got)
	}
	if !strings.Contains(got, "runtime.Barrier()") {
		t.Errorf("expected Barrier call, got:\n%s", got)
	}

	// Barrier must appear after the Parallel call opens (i.e. nested inside).
	par := strings.Index(got, "runtime.Parallel")
	bar := strings.Index(got, "runtime.Barrier")
	if par == -1 || bar == -1 || bar < par {
		t.Errorf("expected Barrier nested inside Parallel, got:\n%s", got)
	}
	// And between the two phases.
	initIdx := strings.Index(got, "initPhase()")
	compIdx := strings.Index(got, "computePhase()")
	if !(initIdx < bar && bar < compIdx) {
		t.Errorf("expected Barrier between initPhase and computePhase; got init=%d bar=%d comp=%d\n%s", initIdx, bar, compIdx, got)
	}
}

// TestTransform_Barrier_AddsRuntimeImport verifies the runtime import is
// injected for a barrier directive in a file with no prior runtime import.
func TestTransform_Barrier_AddsRuntimeImport(t *testing.T) {
	src := `package main

func worker() {
	a()
	//gompher barrier
	b()
}

func a() {}
func b() {}
`
	got := runTransform(t, src)

	if !strings.Contains(got, `"github.com/gomphermp/gomphermp/pkg/runtime"`) {
		t.Errorf("expected runtime import in output, got:\n%s", got)
	}
}

// TestTransform_Barrier_RemovesComment verifies that the //gompher barrier
// comment does not survive in the output, so go/format does not leave it
// dangling near the synthesized call.
func TestTransform_Barrier_RemovesComment(t *testing.T) {
	src := `package main

func worker() {
	a()
	//gompher barrier
	b()
}

func a() {}
func b() {}
`
	got := runTransform(t, src)

	if strings.Contains(got, "//gompher") {
		t.Errorf("expected //gompher barrier comment removed, got:\n%s", got)
	}
}

// TestTransformBarrier_NoContainingBlock verifies the defensive error path:
// when the directive position falls outside any block (here a zero-value
// BarrierDirective whose position is token.NoPos), insertCallAtPos finds no
// enclosing block and transformBarrier returns a descriptive error rather than
// silently dropping the call.
func TestTransformBarrier_NoContainingBlock(t *testing.T) {
	parsed, err := parser.Parse("package main\n\nfunc main() {}\n")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	// Zero-value directive: Pos == token.NoPos, which no real block straddles.
	d := parser.BarrierDirective{}

	err = transformBarrier(parsed, d)
	if err == nil {
		t.Fatal("expected error when barrier has no containing block")
	}
	if !strings.Contains(err.Error(), "could not find containing block") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestTransform_PropagatesBarrierError verifies that a failing transformBarrier
// aborts Transform and returns nil, consistent with the other directives.
func TestTransform_PropagatesBarrierError(t *testing.T) {
	parsed, err := parser.Parse("package main\n\nfunc main() {}\n")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	parsed.Nodes = append(parsed.Nodes, parser.AnnotatedNode{
		Directive: parser.BarrierDirective{},
	})

	got, err := Transform(parsed)
	if err == nil {
		t.Fatal("expected Transform to propagate the underlying error")
	}
	if got != nil {
		t.Errorf("expected nil ParseResult on error, got %v", got)
	}
}
