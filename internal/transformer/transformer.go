// Package transformer rewrites a parsed Go program that contains //gompher
// directives into an equivalent program that calls the gomphermp runtime
// instead.
package transformer

import (
	"github.com/gomphermp/gomphermp/internal/parser"
)

// Transform rewrites every annotated GompherMP directive in the parsed file
// into the corresponding runtime call. Nodes that do not correspond to a
// directive are passed through unmodified.
func Transform(result *parser.ParseResult) (*parser.ParseResult, error) {
	if result == nil {
		return nil, nil
	}

	var emittedRuntime bool
	for _, node := range result.Nodes {
		switch d := node.Directive.(type) {
		case parser.ParallelDirective:
			if err := transformParallel(result, d); err != nil {
				return nil, err
			}
			emittedRuntime = true
		case parser.ForDirective:
			if err := transformFor(result, d); err != nil {
				return nil, err
			}
			emittedRuntime = true
		case parser.ParallelForDirective:
			if err := transformParallelFor(result, d); err != nil {
				return nil, err
			}
			emittedRuntime = true
		case parser.SectionsDirective:
			if err := transformSections(result, d); err != nil {
				return nil, err
			}
			emittedRuntime = true
		case parser.ParallelSectionsDirective:
			if err := transformParallelSections(result, d); err != nil {
				return nil, err
			}
			emittedRuntime = true
		case parser.BarrierDirective:
			if err := transformBarrier(result, d); err != nil {
				return nil, err
			}
			emittedRuntime = true
		case parser.AtomicDirective:
			if err := transformAtomic(result, d); err != nil {
				return nil, err
			}
			emittedRuntime = true
		case parser.CriticalDirective:
			if err := transformCritical(result, d); err != nil {
				return nil, err
			}
			emittedRuntime = true
		case parser.SingleDirective:
			if err := transformSingle(result, d); err != nil {
				return nil, err
			}
			emittedRuntime = true
		case parser.MasterDirective:
			if err := transformMaster(result, d); err != nil {
				return nil, err
			}
			emittedRuntime = true
		case parser.TaskDirective:
			if err := transformTask(result, d); err != nil {
				return nil, err
			}
			emittedRuntime = true
		case parser.TaskwaitDirective:
			if err := transformTaskwait(result, d); err != nil {
				return nil, err
			}
			emittedRuntime = true
		case parser.TaskgroupDirective:
			if err := transformTaskgroup(result, d); err != nil {
				return nil, err
			}
			emittedRuntime = true
		case parser.TaskloopDirective:
			if err := transformTaskloop(result, d); err != nil {
				return nil, err
			}
			emittedRuntime = true
		}
	}

	if emittedRuntime {
		ensureRuntimeImport(result.File)
	}

	return result, nil
}
