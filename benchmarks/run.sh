#!/bin/bash
set -e
GOMPHER=${1:-../gompher}
CSV_OUT=${2:-benchmark_results.csv}
NPROC=$(nproc)
BENCHMARKS="matmul prefixsum mergesort fibonacci sections quicksort pipeline nqueens montecarlo reduce"

echo "═══════════════════════════════════════════════════════════════" >&2
echo "  GompherMP Benchmark Suite  —  seq / manual / gompher" >&2
echo "  CPU: $(grep 'model name' /proc/cpuinfo | head -1 | cut -d: -f2 | xargs)" >&2
echo "  Cores: $NPROC" >&2
echo "═══════════════════════════════════════════════════════════════" >&2
echo "" >&2

for bench in $BENCHMARKS; do
    echo "  Compiling $bench..." >&2
    $GOMPHER build -o /tmp/bench_$bench benchmarks/$bench/main.go
done

echo "benchmark,procs,variant,run,time_ns" > "$CSV_OUT"

echo "" >&2
echo "─── Results  (GOMAXPROCS=$NPROC) ────────────────────────────" >&2
export GOMAXPROCS=$NPROC
for bench in $BENCHMARKS; do
    /tmp/bench_$bench >> "$CSV_OUT"
done

echo "" >&2
echo "─── Scalability sweep ────────────────────────────────────────" >&2
for p in 1 2 4 8 $NPROC; do
    [ $p -gt $NPROC ] && continue
    [ $p -eq $NPROC ] && [ $NPROC -le 8 ] && continue
    echo "" >&2
    echo "  GOMAXPROCS=$p" >&2
    export GOMAXPROCS=$p
    for bench in $BENCHMARKS; do
        /tmp/bench_$bench >> "$CSV_OUT"
    done
done

echo "" >&2
echo "CSV results written to $CSV_OUT" >&2
