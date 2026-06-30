#!/usr/bin/env python3
"""Count parallel-implementation LoC and print benchmark_results.csv rows.

Each output row uses procs=0 as a sentinel for non-timing data:
  benchmark,0,loc_<variant>,0,<line_count>
"""
import os
import re

REPO      = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
BENCH_DIR = os.path.join(REPO, 'benchmarks')

# (benchmark, variant, [func_names_to_sum])
LOC_ENTRIES = [
    ('matmul',      'manual',      ['matMulManual']),
    ('matmul',      'gompher',     ['matMulGompher']),
    ('prefixsum',   'manual',      ['sumManual', 'phasesManual']),
    ('prefixsum',   'gompher',     ['sumGompher', 'phasesGompher']),
    ('mergesort',   'manual',      ['mergeSortManual']),
    ('mergesort',   'gompher',     ['mergeSortGompher']),
    ('heavyreduce', 'manual',      ['sumManual']),
    ('heavyreduce', 'taskloop',    ['sumTaskloop']),
    ('heavyreduce', 'task_depend', ['sumDepend']),
    ('sections',    'manual',      ['processManual']),
    ('sections',    'gompher',     ['processGompher']),
    ('quicksort',   'manual',      ['quickSortManualRec', 'quickSortManual']),
    ('quicksort',   'gompher',     ['quickSortGompher', 'quickSortGompherEntry']),
    ('pipeline',    'manual',      ['pipelineManual']),
    ('pipeline',    'gompher',     ['pipelineGompher']),
    ('nqueens',     'manual',      ['nQueensManual']),
    ('nqueens',     'gompher',     ['nQueensGompher']),
    ('montecarlo',  'manual',      ['monteCarloManual']),
    ('montecarlo',  'gompher',     ['monteCarloGompher']),
    ('reduce',      'manual',      ['reduceManual']),
    ('reduce',      'gompher',     ['reduceGompher']),
]


def count_func(path: str, name: str) -> int:
    pat = re.compile(r'^func ' + re.escape(name) + r'\(')
    in_func, depth, count = False, 0, 0
    with open(path) as fh:
        for line in fh:
            if not in_func:
                if pat.match(line):
                    in_func = True
                else:
                    continue
            count += 1
            depth += line.count('{') - line.count('}')
            if depth == 0:
                break
    return count


for bench, variant, funcs in LOC_ENTRIES:
    src   = os.path.join(BENCH_DIR, bench, 'main.go')
    total = sum(count_func(src, fn) for fn in funcs)
    print(f'{bench},0,loc_{variant},0,{total}')
