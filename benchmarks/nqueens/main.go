package main

import (
	"fmt"
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

func main() {
	runs := 3

	t0 := time.Now()
	var rs int
	for r := 0; r < runs; r++ {
		rs = nQueensSeq()
	}
	tSeq := time.Since(t0) / time.Duration(runs)

	t0 = time.Now()
	var rm int
	for r := 0; r < runs; r++ {
		rm = nQueensManual()
	}
	tMan := time.Since(t0) / time.Duration(runs)

	t0 = time.Now()
	var rg int
	for r := 0; r < runs; r++ {
		rg = nQueensGompher()
	}
	tGmp := time.Since(t0) / time.Duration(runs)

	fmt.Printf("N-Queens\tseq=%v\tmanual=%v\tgompher=%v\tspeedup_manual=%.2fx\tspeedup_gompher=%.2fx\tgmp_vs_manual=%.2fx\tcorrect=%v/%v\tsolutions=%d\n",
		tSeq, tMan, tGmp,
		float64(tSeq)/float64(tMan), float64(tSeq)/float64(tGmp), float64(tMan)/float64(tGmp),
		rs == rm, rs == rg, rs)
}
