package transformer

import (
	"fmt"
	"strings"
	"github.com/gomphermp/gomphermp/internal/parser"
)

// namer produces the synthesized function names that replace the extracted
// body of each directive. Names follow the form "__gompher_<kind>_<n>" where
// n is a per-kind counter that starts at zero on every namer instance.
//
// A counter is preferred over a UUID for three reasons: determinism (the same
// input file always produces the same output, which makes diffing transformer
// output and writing golden tests possible), readability (a stack trace
// pointing at "__gompher_parallel_3" maps trivially to the third parallel
// directive in source order), and reproducibility across runs of the
// transformer on the same file.
type namer struct {
	counters map[parser.DirectiveKind]int
}

func newNamer() *namer {
	return &namer{counters: make(map[parser.DirectiveKind]int)}
}

// next returns the next available identifier for the given directive kind.
// The "__gompher_" prefix follows the same name-mangling convention used by
// C++ compilers for synthesized symbols, keeping generated identifiers well
// outside the namespace that a user would normally write in hand-authored
// Go code.
func (n *namer) next(kind parser.DirectiveKind) string {
	safe := strings.ReplaceAll(string(kind), " ", "_")
	name := fmt.Sprintf("__gompher_%s_%d", safe, n.counters[kind])
	n.counters[kind]++
	return name
}
