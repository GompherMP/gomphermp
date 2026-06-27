package main

import (
	"fmt"
	"math"
	"os"
	goruntime "runtime"
	"sync"
	"time"
)

const SZ = 3 * 1024 * 1024

func numProcs() int { return goruntime.GOMAXPROCS(0) }

func processSeq(data []float64) (float64, float64, float64) {
	n := len(data) / 3
	s0, s1, s2 := 0.0, 0.0, 0.0
	for i := 0; i < n; i++ {
		s0 += math.Sqrt(data[i])
	}
	for i := n; i < 2*n; i++ {
		s1 += data[i] * data[i]
	}
	for i := 2 * n; i < 3*n; i++ {
		s2 += math.Log1p(data[i])
	}
	return s0, s1, s2
}

func processManual(data []float64) (float64, float64, float64) {
	n := len(data) / 3
	var s0, s1, s2 float64
	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		for i := 0; i < n; i++ {
			s0 += math.Sqrt(data[i])
		}
	}()
	go func() {
		defer wg.Done()
		for i := n; i < 2*n; i++ {
			s1 += data[i] * data[i]
		}
	}()
	go func() {
		defer wg.Done()
		for i := 2 * n; i < 3*n; i++ {
			s2 += math.Log1p(data[i])
		}
	}()
	wg.Wait()
	return s0, s1, s2
}

func processGompher(data []float64) (float64, float64, float64) {
	n := len(data) / 3
	s0, s1, s2 := 0.0, 0.0, 0.0
	//gompher parallel sections reduction(+:s0) reduction(+:s1) reduction(+:s2)
	{
		//gompher section
		{
			for i := 0; i < n; i++ {
				s0 += math.Sqrt(data[i])
			}
		}
		//gompher section
		{
			for i := n; i < 2*n; i++ {
				s1 += data[i] * data[i]
			}
		}
		//gompher section
		{
			for i := 2 * n; i < 3*n; i++ {
				s2 += math.Log1p(data[i])
			}
		}
	}
	return s0, s1, s2
}

func criticalDemo() int {
	counter := 0
	//gompher parallel
	{
		//gompher critical
		{
			counter++
		}
	}
	return counter
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
	data := make([]float64, SZ)
	for i := range data {
		data[i] = float64(i%1000) + 1.0
	}
	const runs = 10
	timesSeq := make([]time.Duration, runs)
	timesMan := make([]time.Duration, runs)
	timesGmp := make([]time.Duration, runs)

	var a0s, b0s, c0s float64
	var a0m, b0m, c0m float64
	var a0g, b0g, c0g float64

	for r := 0; r < runs; r++ {
		t0 := time.Now()
		a0s, b0s, c0s = processSeq(data)
		timesSeq[r] = time.Since(t0)
	}
	for r := 0; r < runs; r++ {
		t0 := time.Now()
		a0m, b0m, c0m = processManual(data)
		timesMan[r] = time.Since(t0)
	}
	for r := 0; r < runs; r++ {
		t0 := time.Now()
		a0g, b0g, c0g = processGompher(data)
		timesGmp[r] = time.Since(t0)
	}

	eps := 1e-6
	okM := math.Abs(a0s-a0m)/a0s < eps && math.Abs(b0s-b0m)/b0s < eps && math.Abs(c0s-c0m)/c0s < eps
	okG := math.Abs(a0s-a0g)/a0s < eps && math.Abs(b0s-b0g)/b0s < eps && math.Abs(c0s-c0g)/c0s < eps

	counter := criticalDemo()

	printCSV("sections", "seq", timesSeq)
	printCSV("sections", "manual", timesMan)
	printCSV("sections", "gompher", timesGmp)

	tSeq, tSeqStd := durMean(timesSeq), durStd(timesSeq, durMean(timesSeq))
	tMan, tManStd := durMean(timesMan), durStd(timesMan, durMean(timesMan))
	tGmp, tGmpStd := durMean(timesGmp), durStd(timesGmp, durMean(timesGmp))

	fmt.Fprintf(os.Stderr, "Sections\tseq=%v±%v\tmanual=%v±%v\tgompher=%v±%v\tspeedup_manual=%.2fx\tspeedup_gompher=%.2fx\tgmp_vs_manual=%.2fx\tcorrect=%v/%v\tcounter=%d(=%d cores)\n",
		tSeq, tSeqStd, tMan, tManStd, tGmp, tGmpStd,
		float64(tSeq)/float64(tMan), float64(tSeq)/float64(tGmp), float64(tMan)/float64(tGmp),
		okM, okG, counter, numProcs())
}
