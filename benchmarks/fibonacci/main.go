package main

import (
	"fmt"
	"math"
	goruntime "runtime"
	"time"
)

const VecN = 1 << 20

func numProcs() int { return goruntime.GOMAXPROCS(0) }

func heavySqrtSum(data []float64, lo, hi int) float64 {
	s := 0.0
	for i := lo; i < hi; i++ { s += math.Sqrt(data[i]) }
	return s
}

func sumSeq(data []float64) float64 {
	s := 0.0; for _, v := range data { s += math.Sqrt(v) }; return s
}

func sumManual(data []float64) float64 {
	p, chunk := numProcs(), len(data)/numProcs()
	ch := make(chan float64, p)
	for t := 0; t < p; t++ {
		lo, hi := t*chunk, (t+1)*chunk
		if t == p-1 { hi = len(data) }
		go func(lo, hi int) { ch <- heavySqrtSum(data, lo, hi) }(lo, hi)
	}
	total := 0.0
	for range p { total += <-ch }
	return total
}

func sumDepend(data []float64) float64 {
	n := len(data); q := n / 4
	var r0, r1, r2, r3 float64
	var m0, m1 float64
	var result float64
	//gompher taskgroup
	{
		//gompher task depend(out:r0)
		{ r0 = heavySqrtSum(data, 0, q) }
		//gompher task depend(out:r1)
		{ r1 = heavySqrtSum(data, q, 2*q) }
		//gompher task depend(out:r2)
		{ r2 = heavySqrtSum(data, 2*q, 3*q) }
		//gompher task depend(out:r3)
		{ r3 = heavySqrtSum(data, 3*q, n) }
		//gompher task depend(in:r0, r1) depend(out:m0)
		{ m0 = r0 + r1 }
		//gompher task depend(in:r2, r3) depend(out:m1)
		{ m1 = r2 + r3 }
		//gompher task depend(in:m0, m1) depend(out:result)
		{ result = m0 + m1 }
	}
	return result
}

func sumTaskloop(data []float64) float64 {
	p := numProcs()
	results := make([]float64, p)
	chunkSize := len(data) / p
	//gompher taskgroup
	{
		//gompher taskloop grainsize(1)
		for c := 0; c < p; c++ {
			lo, hi := c*chunkSize, (c+1)*chunkSize
			if c == p-1 { hi = len(data) }
			results[c] = heavySqrtSum(data, lo, hi)
		}
	}
	total := 0.0; for _, v := range results { total += v }; return total
}

func main() {
	data := make([]float64, VecN)
	for i := range data { data[i] = float64(i + 1) }
	runs := 5

	t0 := time.Now()
	var rs float64
	for r := 0; r < runs; r++ { rs = sumSeq(data) }
	tSeq := time.Since(t0) / time.Duration(runs)

	t0 = time.Now()
	var rm float64
	for r := 0; r < runs; r++ { rm = sumManual(data) }
	tMan := time.Since(t0) / time.Duration(runs)

	t0 = time.Now()
	var rd float64
	for r := 0; r < runs; r++ { rd = sumDepend(data) }
	tDep := time.Since(t0) / time.Duration(runs)

	t0 = time.Now()
	var rl float64
	for r := 0; r < runs; r++ { rl = sumTaskloop(data) }
	tLoop := time.Since(t0) / time.Duration(runs)

	eps := rs * 1e-6
	fmt.Printf("B5 Task depend\tseq=%v\tmanual=%v\tgompher=%v\tspeedup_manual=%.2fx\tspeedup_gompher=%.2fx\tgmp_vs_manual=%.2fx\tcorrect=%v/%v\n",
		tSeq, tMan, tDep,
		float64(tSeq)/float64(tMan), float64(tSeq)/float64(tDep), float64(tMan)/float64(tDep),
		math.Abs(rm-rs) < eps, math.Abs(rd-rs) < eps)
	fmt.Printf("B5 Taskloop\tseq=%v\tmanual=%v\tgompher=%v\tspeedup_manual=%.2fx\tspeedup_gompher=%.2fx\tgmp_vs_manual=%.2fx\tcorrect=%v/%v\n",
		tSeq, tMan, tLoop,
		float64(tSeq)/float64(tMan), float64(tSeq)/float64(tLoop), float64(tMan)/float64(tLoop),
		math.Abs(rm-rs) < eps, math.Abs(rl-rs) < eps)
}
