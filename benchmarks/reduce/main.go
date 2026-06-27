package main

import (
	"fmt"
	"math"
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

func main() {
	runs := 5

	t0 := time.Now()
	var rs float64
	for r := 0; r < runs; r++ {
		rs = reduceSeq()
	}
	tSeq := time.Since(t0) / time.Duration(runs)

	t0 = time.Now()
	var rm float64
	for r := 0; r < runs; r++ {
		rm = reduceManual()
	}
	tMan := time.Since(t0) / time.Duration(runs)

	t0 = time.Now()
	var rg float64
	for r := 0; r < runs; r++ {
		rg = reduceGompher()
	}
	tGmp := time.Since(t0) / time.Duration(runs)

	eps := math.Abs(rs) * 1e-9
	fmt.Printf("Reduce\tseq=%v\tmanual=%v\tgompher=%v\tspeedup_manual=%.2fx\tspeedup_gompher=%.2fx\tgmp_vs_manual=%.2fx\tcorrect=%v/%v\n",
		tSeq, tMan, tGmp,
		float64(tSeq)/float64(tMan), float64(tSeq)/float64(tGmp), float64(tMan)/float64(tGmp),
		math.Abs(rm-rs) < eps, math.Abs(rg-rs) < eps)
}
