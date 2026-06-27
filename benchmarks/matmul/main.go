package main

import (
	"fmt"
	"math"
	"os"
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
	A, B := newMatrix(), newMatrix()
	Cs := make([][]float64, N)
	Cm := make([][]float64, N)
	Cg := make([][]float64, N)
	for i := range Cs {
		Cs[i] = make([]float64, N)
		Cm[i] = make([]float64, N)
		Cg[i] = make([]float64, N)
	}

	const runs = 10
	timesSeq := make([]time.Duration, runs)
	timesMan := make([]time.Duration, runs)
	timesGmp := make([]time.Duration, runs)

	for r := 0; r < runs; r++ {
		t0 := time.Now()
		matMulSeq(A, B, Cs)
		timesSeq[r] = time.Since(t0)
	}
	for r := 0; r < runs; r++ {
		t0 := time.Now()
		matMulManual(A, B, Cm)
		timesMan[r] = time.Since(t0)
	}
	for r := 0; r < runs; r++ {
		t0 := time.Now()
		matMulGompher(A, B, Cg)
		timesGmp[r] = time.Since(t0)
	}

	printCSV("matmul", "seq", timesSeq)
	printCSV("matmul", "manual", timesMan)
	printCSV("matmul", "gompher", timesGmp)

	tSeq, tSeqStd := durMean(timesSeq), durStd(timesSeq, durMean(timesSeq))
	tMan, tManStd := durMean(timesMan), durStd(timesMan, durMean(timesMan))
	tGmp, tGmpStd := durMean(timesGmp), durStd(timesGmp, durMean(timesGmp))

	fmt.Fprintf(os.Stderr, "MatMul\tseq=%v±%v\tmanual=%v±%v\tgompher=%v±%v\tspeedup_manual=%.2fx\tspeedup_gompher=%.2fx\tgmp_vs_manual=%.2fx\tcorrect=%v/%v\n",
		tSeq, tSeqStd, tMan, tManStd, tGmp, tGmpStd,
		float64(tSeq)/float64(tMan), float64(tSeq)/float64(tGmp), float64(tMan)/float64(tGmp),
		equal(Cs, Cm), equal(Cs, Cg))
}
