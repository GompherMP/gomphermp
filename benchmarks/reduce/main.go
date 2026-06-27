package main

import (
	"fmt"
	"math"
	"os"
	goruntime "runtime"
	"sync"
	"time"
)

const (
	N       = 1 << 16
	MaxWork = 512
)

func numProcs() int { return goruntime.GOMAXPROCS(0) }

// compute returns a value for element i with work proportional to i
// (linearly increasing from 0 to MaxWork ops), creating load imbalance
// that benefits from schedule(dynamic) over schedule(static).
func compute(i int) float64 {
	s := math.Sqrt(float64(i + 1))
	ops := i * MaxWork / N
	for k := 0; k < ops; k++ {
		s += math.Sin(s * 0.001)
	}
	return s
}

func reduceSeq() float64 {
	m := -math.MaxFloat64
	for i := 0; i < N; i++ {
		if v := compute(i); v > m {
			m = v
		}
	}
	return m
}

func reduceManual() float64 {
	p := numProcs()
	partials := make([]float64, p)
	for i := range partials {
		partials[i] = -math.MaxFloat64
	}
	chunk := N / p
	var wg sync.WaitGroup
	for t := 0; t < p; t++ {
		s, e := t*chunk, (t+1)*chunk
		if t == p-1 {
			e = N
		}
		wg.Add(1)
		go func(t, s, e int) {
			defer wg.Done()
			m := -math.MaxFloat64
			for i := s; i < e; i++ {
				if v := compute(i); v > m {
					m = v
				}
			}
			partials[t] = m
		}(t, s, e)
	}
	wg.Wait()
	m := -math.MaxFloat64
	for _, v := range partials {
		if v > m {
			m = v
		}
	}
	return m
}

func reduceGompher() float64 {
	m := -math.MaxFloat64
	//gompher parallel for schedule(dynamic, 64) reduction(max:m)
	for i := 0; i < N; i++ {
		if v := compute(i); v > m {
			m = v
		}
	}
	return m
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
		rs = reduceSeq()
		timesSeq[r] = time.Since(t0)
	}
	for r := 0; r < runs; r++ {
		t0 := time.Now()
		rm = reduceManual()
		timesMan[r] = time.Since(t0)
	}
	for r := 0; r < runs; r++ {
		t0 := time.Now()
		rg = reduceGompher()
		timesGmp[r] = time.Since(t0)
	}

	printCSV("reduce", "seq", timesSeq)
	printCSV("reduce", "manual", timesMan)
	printCSV("reduce", "gompher", timesGmp)

	eps := math.Abs(rs) * 1e-9
	tSeq, tSeqStd := durMean(timesSeq), durStd(timesSeq, durMean(timesSeq))
	tMan, tManStd := durMean(timesMan), durStd(timesMan, durMean(timesMan))
	tGmp, tGmpStd := durMean(timesGmp), durStd(timesGmp, durMean(timesGmp))

	fmt.Fprintf(os.Stderr, "Reduce\tseq=%v±%v\tmanual=%v±%v\tgompher=%v±%v\tspeedup_manual=%.2fx\tspeedup_gompher=%.2fx\tgmp_vs_manual=%.2fx\tcorrect=%v/%v\n",
		tSeq, tSeqStd, tMan, tManStd, tGmp, tGmpStd,
		float64(tSeq)/float64(tMan), float64(tSeq)/float64(tGmp), float64(tMan)/float64(tGmp),
		math.Abs(rm-rs) < eps, math.Abs(rg-rs) < eps)
}
