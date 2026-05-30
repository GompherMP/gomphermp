package transformer

import (
	"strings"
	"testing"
	"github.com/gomphermp/gomphermp/internal/parser"
)

// TestNamer_ProducesUniquePerKind verifies that repeated invocations for the
// same directive kind yield distinct identifiers. This is the essential
// property: if two extracted bodies got the same name, the generated Go
// file would have duplicate declarations and fail to compile.
func TestNamer_ProducesUniquePerKind(t *testing.T) {
	n := newNamer()
	a := n.next(parser.DirParallel)
	b := n.next(parser.DirParallel)
	if a == b {
		t.Errorf("expected distinct names, got both = %q", a)
	}
}

// TestNamer_IndependentCountersPerKind verifies that the counters for
// different kinds advance independently, so naming a "for" does not consume
// a slot in the "parallel" sequence.
func TestNamer_IndependentCountersPerKind(t *testing.T) {
	n := newNamer()
	p := n.next(parser.DirParallel)
	f := n.next(parser.DirFor)
	if !strings.Contains(p, "parallel_0") {
		t.Errorf("expected parallel_0 in %q", p)
	}
	if !strings.Contains(f, "for_0") {
		t.Errorf("expected for_0 in %q", f)
	}
}

// TestNamer_DeterministicAcrossInstances verifies that two fresh namer
// instances produce the same sequence for the same kinds. This is a
// precondition for reproducible transformer output and stable golden tests.
func TestNamer_DeterministicAcrossInstances(t *testing.T) {
	a := newNamer()
	b := newNamer()
	kinds := []parser.DirectiveKind{parser.DirParallel, parser.DirFor, parser.DirParallel}
	for _, kind := range kinds {
		if a.next(kind) != b.next(kind) {
			t.Fatal("namer output diverged between equivalent instances")
		}
	}
}

// TestNamer_SanitizesMultiWordKinds verifies that kinds containing spaces
// (such as "parallel for") are converted into valid Go identifiers by
// replacing the space with an underscore.
func TestNamer_SanitizesMultiWordKinds(t *testing.T) {
	n := newNamer()
	got := n.next(parser.DirParallelFor)
	if strings.Contains(got, " ") {
		t.Errorf("identifier contains space: %q", got)
	}
	if !strings.Contains(got, "parallel_for") {
		t.Errorf("expected sanitized kind in %q", got)
	}
}

// TestNamer_PrefixIsApplied verifies that generated identifiers carry the
// "__gompher_" prefix that distinguishes synthesized code from anything a
// user is likely to write by hand.
func TestNamer_PrefixIsApplied(t *testing.T) {
	n := newNamer()
	got := n.next(parser.DirParallel)
	if !strings.HasPrefix(got, "__gompher_") {
		t.Errorf("expected __gompher_ prefix, got %q", got)
	}
}
