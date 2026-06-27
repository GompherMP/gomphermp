package main

import (
	"fmt"
	goruntime "runtime"
	"sync"
	"time"
)

const N = 512

func numProcs() int { return goruntime.GOMAXPROCS(0) }

func newMatrix() [][]float64 {
	m := make([][]float64, N)
	for i := range m {
		m[i] = make([]float64, N)
		for j := range m[i] {
			m[i][j] = float64(i*N+j) * 0.001
		}
	}
	return m
}

func matMulSeq(A, B, C [][]float64) {
	for i := 0; i < N; i++ {
		for j := 0; j < N; j++ {
			sum := 0.0
			for k := 0; k < N; k++ {
				sum += A[i][k] * B[k][j]
			}
			C[i][j] = sum
		}
	}
}

func matMulManual(A, B, C [][]float64) {
	p, chunk := numProcs(), N/numProcs()
	var wg sync.WaitGroup
	for t := 0; t < p; t++ {
		s, e := t*chunk, (t+1)*chunk
		if t == p-1 {
			e = N
		}
		wg.Add(1)
		go func(s, e int) {
			defer wg.Done()
			for i := s; i < e; i++ {
				for j := 0; j < N; j++ {
					sum := 0.0
					for k := 0; k < N; k++ {
						sum += A[i][k] * B[k][j]
					}
					C[i][j] = sum
				}
			}
		}(s, e)
	}
	wg.Wait()
}

func matMulGompher(A, B, C [][]float64) {
	//gompher parallel for schedule(static, 16)
	for i := 0; i < N; i++ {
		for j := 0; j < N; j++ {
			sum := 0.0
			for k := 0; k < N; k++ {
				sum += A[i][k] * B[k][j]
			}
			C[i][j] = sum
		}
	}
}

func equal(A, B [][]float64) bool {
	for i := 0; i < N; i++ {
		for j := 0; j < N; j++ {
			d := A[i][j] - B[i][j]
			if d < -1e-9 || d > 1e-9 {
				return false
			}
		}
	}
	return true
}

func main() {
	A, B := newMatrix(), newMatrix()
	Cs := make([][]float64, N)
	Cm := make([][]float64, N)
	Cg := make([][]float64, N)
	for i := range Cs {
		Cs[i] = make([]float64, N)
		Cm[i] = make([]float64, N)
		Cg[i] = make([]float64, N)
	}

	runs := 5
	t0 := time.Now()
	for r := 0; r < runs; r++ {
		matMulSeq(A, B, Cs)
	}
	tSeq := time.Since(t0) / time.Duration(runs)

	t0 = time.Now()
	for r := 0; r < runs; r++ {
		matMulManual(A, B, Cm)
	}
	tMan := time.Since(t0) / time.Duration(runs)

	t0 = time.Now()
	for r := 0; r < runs; r++ {
		matMulGompher(A, B, Cg)
	}
	tGmp := time.Since(t0) / time.Duration(runs)

	fmt.Printf("MatMul\tseq=%v\tmanual=%v\tgompher=%v\tspeedup_manual=%.2fx\tspeedup_gompher=%.2fx\tgmp_vs_manual=%.2fx\tcorrect=%v/%v\n",
		tSeq, tMan, tGmp,
		float64(tSeq)/float64(tMan), float64(tSeq)/float64(tGmp), float64(tMan)/float64(tGmp),
		equal(Cs, Cm), equal(Cs, Cg))
}
