#!/bin/bash
set -e
GOMPHER=${1:-../gompher}
NPROC=$(nproc)
BENCHMARKS="matmul prefixsum mergesort fibonacci sections quicksort pipeline nqueens montecarlo reduce"

echo "═══════════════════════════════════════════════════════════════"
echo "  GompherMP Benchmark Suite  —  seq / manual / gompher"
echo "  CPU: $(grep 'model name' /proc/cpuinfo | head -1 | cut -d: -f2 | xargs)"
echo "  Cores: $NPROC"
echo "═══════════════════════════════════════════════════════════════"
echo ""

for bench in $BENCHMARKS; do
    echo "  Compiling $bench..."
    $GOMPHER build -o /tmp/bench_$bench benchmarks/$bench/main.go
done

echo ""
echo "─── Results  (GOMAXPROCS=$NPROC) ────────────────────────────"
export GOMAXPROCS=$NPROC
for bench in $BENCHMARKS; do
    /tmp/bench_$bench
done

echo ""
echo "─── Scalability sweep ────────────────────────────────────────"
for p in 1 2 4 8 $NPROC; do
    [ $p -gt $NPROC ] && continue
    [ $p -eq $NPROC ] && [ $NPROC -le 8 ] && continue
    echo ""
    echo "  GOMAXPROCS=$p"
    export GOMAXPROCS=$p
    for bench in $BENCHMARKS; do
        /tmp/bench_$bench
    done
done
