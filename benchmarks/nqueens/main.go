package main

import (
	"fmt"
	"math"
	"os"
	goruntime "runtime"
	"time"
)

const BoardN = 15

func numProcs() int { return goruntime.GOMAXPROCS(0) }

func solve(row, cols, diag1, diag2 int) int {
	if row == BoardN {
		return 1
	}
	count := 0
	available := ((1 << BoardN) - 1) &^ (cols | diag1 | diag2)
	for available != 0 {
		bit := available & (-available)
		available &= available - 1
		count += solve(row+1, cols|bit, (diag1|bit)>>1, (diag2|bit)<<1)
	}
	return count
}

func nQueensSeq() int {
	return solve(0, 0, 0, 0)
}

func nQueensManual() int {
	results := make(chan int, BoardN)
	for col := 0; col < BoardN; col++ {
		bit := 1 << col
		go func(bit int) {
			results <- solve(1, bit, bit>>1, bit<<1)
		}(bit)
	}
	total := 0
	for i := 0; i < BoardN; i++ {
		total += <-results
	}
	return total
}

func nQueensGompher() int {
	results := make([]int, BoardN)
	//gompher taskgroup
	{
		for col := 0; col < BoardN; col++ {
			bit := 1 << col
			c := col
			//gompher task firstprivate(bit, c)
			{
				results[c] = solve(1, bit, bit>>1, bit<<1)
			}
		}
	}
	total := 0
	for _, v := range results {
		total += v
	}
	return total
}

func durMean(ts []time.Duration) time.Duration {
	var sum int64
	for _, t := range ts {
		sum += int64(t)
	}
	return time.Duration(sum / int64(len(ts)))
}

func durStd(ts []time.Duration, m time.Duration) time.Duration {
	var v float64
	for _, t := range ts {
		d := float64(int64(t) - int64(m))
		v += d * d
	}
	return time.Duration(math.Sqrt(v / float64(len(ts))))
}

func printCSV(name, variant string, times []time.Duration) {
	p := numProcs()
	for i, t := range times {
		fmt.Printf("%s,%d,%s,%d,%d\n", name, p, variant, i+1, int64(t))
	}
}

func main() {
	const runs = 10
	timesSeq := make([]time.Duration, runs)
	timesMan := make([]time.Duration, runs)
	timesGmp := make([]time.Duration, runs)

	var rs, rm, rg int
	for r := 0; r < runs; r++ {
		t0 := time.Now()
		rs = nQueensSeq()
		timesSeq[r] = time.Since(t0)
	}
	for r := 0; r < runs; r++ {
		t0 := time.Now()
		rm = nQueensManual()
		timesMan[r] = time.Since(t0)
	}
	for r := 0; r < runs; r++ {
		t0 := time.Now()
		rg = nQueensGompher()
		timesGmp[r] = time.Since(t0)
	}

	printCSV("nqueens", "seq", timesSeq)
	printCSV("nqueens", "manual", timesMan)
	printCSV("nqueens", "gompher", timesGmp)

	tSeq, tSeqStd := durMean(timesSeq), durStd(timesSeq, durMean(timesSeq))
	tMan, tManStd := durMean(timesMan), durStd(timesMan, durMean(timesMan))
	tGmp, tGmpStd := durMean(timesGmp), durStd(timesGmp, durMean(timesGmp))

	fmt.Fprintf(os.Stderr, "N-Queens\tseq=%v±%v\tmanual=%v±%v\tgompher=%v±%v\tspeedup_manual=%.2fx\tspeedup_gompher=%.2fx\tgmp_vs_manual=%.2fx\tcorrect=%v/%v\tsolutions=%d\n",
		tSeq, tSeqStd, tMan, tManStd, tGmp, tGmpStd,
		float64(tSeq)/float64(tMan), float64(tSeq)/float64(tGmp), float64(tMan)/float64(tGmp),
		rs == rm, rs == rg, rs)
}
