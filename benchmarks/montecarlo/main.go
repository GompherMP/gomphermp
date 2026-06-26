package main

import (
	"fmt"
	"math/rand"
	goruntime "runtime"
	"time"
)

const Samples = 10_000_000

func numProcs() int { return goruntime.GOMAXPROCS(0) }

func piSeq() float64 {
	hits := 0
	rng := rand.New(rand.NewSource(42))
	for i := 0; i < Samples; i++ {
		x, y := rng.Float64(), rng.Float64()
		if x*x+y*y <= 1.0 { hits++ }
	}
	return 4.0 * float64(hits) / float64(Samples)
}

func piManual() float64 {
	xs := make([]float64, Samples)
	ys := make([]float64, Samples)
	rng := rand.New(rand.NewSource(42))
	for i := range xs { xs[i], ys[i] = rng.Float64(), rng.Float64() }
	p, per := numProcs(), Samples/numProcs()
	ch := make(chan int64, p)
	for t := 0; t < p; t++ {
		s, e := t*per, (t+1)*per
		if t == p-1 { e = Samples }
		go func(s, e int) {
			var n int64
			for i := s; i < e; i++ {
				if xs[i]*xs[i]+ys[i]*ys[i] <= 1.0 { n++ }
			}
			ch <- n
		}(s, e)
	}
	total := int64(0)
	for range p { total += <-ch }
	return 4.0 * float64(total) / float64(Samples)
}

func piGompher() float64 {
	xs := make([]float64, Samples)
	ys := make([]float64, Samples)
	rng := rand.New(rand.NewSource(42))
	for i := 0; i < Samples; i++ { xs[i] = rng.Float64(); ys[i] = rng.Float64() }
	hits := 0
	//gompher parallel for reduction(+:hits)
	for i := 0; i < Samples; i++ {
		if xs[i]*xs[i]+ys[i]*ys[i] <= 1.0 { hits++ }
	}
	return 4.0 * float64(hits) / float64(Samples)
}

func main() {
	runs := 5
	t0 := time.Now()
	var rs float64
	for r := 0; r < runs; r++ { rs = piSeq() }
	tSeq := time.Since(t0) / time.Duration(runs)

	t0 = time.Now()
	var rm float64
	for r := 0; r < runs; r++ { rm = piManual() }
	tMan := time.Since(t0) / time.Duration(runs)

	t0 = time.Now()
	var rg float64
	for r := 0; r < runs; r++ { rg = piGompher() }
	tGmp := time.Since(t0) / time.Duration(runs)

	fmt.Printf("B2 Montecarlo\tseq=%v\tmanual=%v\tgompher=%v\tspeedup_manual=%.2fx\tspeedup_gompher=%.2fx\tgmp_vs_manual=%.2fx\tpi=%.5f/%.5f/%.5f\n",
		tSeq, tMan, tGmp,
		float64(tSeq)/float64(tMan), float64(tSeq)/float64(tGmp), float64(tMan)/float64(tGmp),
		rs, rm, rg)
}
