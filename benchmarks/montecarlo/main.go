package main

import (
	"fmt"
	"math"
	"os"
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

	var rs, rm, rg float64
	for r := 0; r < runs; r++ {
		t0 := time.Now()
		rs = monteCarloSeq()
		timesSeq[r] = time.Since(t0)
	}
	for r := 0; r < runs; r++ {
		t0 := time.Now()
		rm = monteCarloManual()
		timesMan[r] = time.Since(t0)
	}
	for r := 0; r < runs; r++ {
		t0 := time.Now()
		rg = monteCarloGompher()
		timesGmp[r] = time.Since(t0)
	}

	printCSV("montecarlo", "seq", timesSeq)
	printCSV("montecarlo", "manual", timesMan)
	printCSV("montecarlo", "gompher", timesGmp)

	eps := 0.01
	tSeq, tSeqStd := durMean(timesSeq), durStd(timesSeq, durMean(timesSeq))
	tMan, tManStd := durMean(timesMan), durStd(timesMan, durMean(timesMan))
	tGmp, tGmpStd := durMean(timesGmp), durStd(timesGmp, durMean(timesGmp))

	fmt.Fprintf(os.Stderr, "MonteCarlo\tseq=%v±%v\tmanual=%v±%v\tgompher=%v±%v\tspeedup_manual=%.2fx\tspeedup_gompher=%.2fx\tgmp_vs_manual=%.2fx\tpi_err_manual=%.6f\tpi_err_gompher=%.6f\tcorrect=%v/%v\n",
		tSeq, tSeqStd, tMan, tManStd, tGmp, tGmpStd,
		float64(tSeq)/float64(tMan), float64(tSeq)/float64(tGmp), float64(tMan)/float64(tGmp),
		math.Abs(rm-math.Pi), math.Abs(rg-math.Pi),
		math.Abs(rm-rs) < eps, math.Abs(rg-rs) < eps)
}
