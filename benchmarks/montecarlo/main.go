package main

import (
	"fmt"
	"math"
	goruntime "runtime"
	"time"
)

const M = 1 << 24

func numProcs() int { return goruntime.GOMAXPROCS(0) }

// sample generates a deterministic pseudo-random (x, y) in [0,1)^2 via LCG.
func sample(i int) (float64, float64) {
	h := uint64(i)*6364136223846793005 + 1442695040888963407
	x := float64(h>>11) / float64(1<<53)
	h = h*6364136223846793005 + 1442695040888963407
	y := float64(h>>11) / float64(1<<53)
	return x, y
}

func monteCarloSeq() float64 {
	hits := 0
	for i := 0; i < M; i++ {
		x, y := sample(i)
		if x*x+y*y <= 1.0 {
			hits++
		}
	}
	return 4.0 * float64(hits) / float64(M)
}

func monteCarloManual() float64 {
	p := numProcs()
	chunk := M / p
	ch := make(chan int, p)
	for t := 0; t < p; t++ {
		s, e := t*chunk, (t+1)*chunk
		if t == p-1 {
			e = M
		}
		go func(s, e int) {
			h := 0
			for i := s; i < e; i++ {
				x, y := sample(i)
				if x*x+y*y <= 1.0 {
					h++
				}
			}
			ch <- h
		}(s, e)
	}
	total := 0
	for range p {
		total += <-ch
	}
	return 4.0 * float64(total) / float64(M)
}

func monteCarloGompher() float64 {
	hits := 0
	//gompher parallel for schedule(static, 1024) reduction(+:hits)
	for i := 0; i < M; i++ {
		x, y := sample(i)
		if x*x+y*y <= 1.0 {
			hits++
		}
	}
	return 4.0 * float64(hits) / float64(M)
}

func main() {
	runs := 5

	t0 := time.Now()
	var rs float64
	for r := 0; r < runs; r++ {
		rs = monteCarloSeq()
	}
	tSeq := time.Since(t0) / time.Duration(runs)

	t0 = time.Now()
	var rm float64
	for r := 0; r < runs; r++ {
		rm = monteCarloManual()
	}
	tMan := time.Since(t0) / time.Duration(runs)

	t0 = time.Now()
	var rg float64
	for r := 0; r < runs; r++ {
		rg = monteCarloGompher()
	}
	tGmp := time.Since(t0) / time.Duration(runs)

	eps := 0.01
	fmt.Printf("MonteCarlo\tseq=%v\tmanual=%v\tgompher=%v\tspeedup_manual=%.2fx\tspeedup_gompher=%.2fx\tgmp_vs_manual=%.2fx\tpi_err_manual=%.6f\tpi_err_gompher=%.6f\tcorrect=%v/%v\n",
		tSeq, tMan, tGmp,
		float64(tSeq)/float64(tMan), float64(tSeq)/float64(tGmp), float64(tMan)/float64(tGmp),
		math.Abs(rm-math.Pi), math.Abs(rg-math.Pi),
		math.Abs(rm-rs) < eps, math.Abs(rg-rs) < eps)
}
