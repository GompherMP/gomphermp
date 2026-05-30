package transformer

import (
	"fmt"
	"strings"
	"github.com/gomphermp/gomphermp/internal/parser"
)

// namer produces the synthesized function names that replace the extracted
// body of each directive. Names follow the form "__gompher_<kind>_<n>" where
// n is a per-kind counter that starts at zero on every namer instance.
type namer struct {
	counters map[parser.DirectiveKind]int
}

func newNamer() *namer {
	return &namer{counters: make(map[parser.DirectiveKind]int)}
}

// next returns the next available identifier for the given directive kind.
func (n *namer) next(kind parser.DirectiveKind) string {
	safe := strings.ReplaceAll(string(kind), " ", "_")
	name := fmt.Sprintf("__gompher_%s_%d", safe, n.counters[kind])
	n.counters[kind]++
	return name
}
