package main

import (
	"fmt"
	"math"
	"os"
	goruntime "runtime"
	"sync"
	"time"
)

const N = 4 * 1024 * 1024

func numProcs() int { return goruntime.GOMAXPROCS(0) }

func sumSeq(data []float64) float64 {
	total := 0.0
	for _, v := range data {
		total += v
	}
	return total
}

func sumManual(data []float64) float64 {
	p, chunk := numProcs(), len(data)/numProcs()
	ch := make(chan float64, p)
	for t := 0; t < p; t++ {
		s, e := t*chunk, (t+1)*chunk
		if t == p-1 {
			e = len(data)
		}
		go func(s, e int) {
			local := 0.0
			for _, v := range data[s:e] {
				local += v
			}
			ch <- local
		}(s, e)
	}
	total := 0.0
	for range p {
		total += <-ch
	}
	return total
}

func sumGompher(data []float64) float64 {
	total := 0.0
	//gompher parallel for reduction(+:total)
	for i := 0; i < len(data); i++ {
		total += data[i]
	}
	return total
}

func phasesManual(data []int) int {
	p, chunk := numProcs(), len(data)/numProcs()
	var wg sync.WaitGroup
	for t := 0; t < p; t++ {
		s, e := t*chunk, (t+1)*chunk
		if t == p-1 {
			e = len(data)
		}
		wg.Add(1)
		go func(s, e int) {
			defer wg.Done()
			for i := s; i < e; i++ {
				data[i] *= 2
			}
		}(s, e)
	}
	wg.Wait()
	data[0] = -999
	for t := 0; t < p; t++ {
		s, e := t*chunk, (t+1)*chunk
		if t == p-1 {
			e = len(data)
		}
		wg.Add(1)
		go func(s, e int) {
			defer wg.Done()
			for i := s; i < e; i++ {
				data[i] += 1
			}
		}(s, e)
	}
	wg.Wait()
	return data[0]
}

func phasesGompher(data []int) int {
	//gompher parallel
	{
		//gompher for
		for i := 0; i < len(data); i++ {
			data[i] *= 2
		}
		//gompher single
		{
			data[0] = -999
		}
		//gompher barrier
		//gompher for
		for i := 0; i < len(data); i++ {
			data[i] += 1
		}
	}
	return data[0]
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
	data := make([]float64, N)
	for i := range data {
		data[i] = 1.0
	}
	const runs = 10
	timesSeq := make([]time.Duration, runs)
	timesMan := make([]time.Duration, runs)
	timesGmp := make([]time.Duration, runs)

	var rs, rm, rg float64
	for r := 0; r < runs; r++ {
		t0 := time.Now()
		rs = sumSeq(data)
		timesSeq[r] = time.Since(t0)
	}
	for r := 0; r < runs; r++ {
		t0 := time.Now()
		rm = sumManual(data)
		timesMan[r] = time.Since(t0)
	}
	for r := 0; r < runs; r++ {
		t0 := time.Now()
		rg = sumGompher(data)
		timesGmp[r] = time.Since(t0)
	}

	small := make([]int, 100)
	for i := range small {
		small[i] = i
	}
	sentM := phasesManual(small)
	for i := range small {
		small[i] = i
	}
	sentG := phasesGompher(small)

	printCSV("prefixsum", "seq", timesSeq)
	printCSV("prefixsum", "manual", timesMan)
	printCSV("prefixsum", "gompher", timesGmp)

	tSeq, tSeqStd := durMean(timesSeq), durStd(timesSeq, durMean(timesSeq))
	tMan, tManStd := durMean(timesMan), durStd(timesMan, durMean(timesMan))
	tGmp, tGmpStd := durMean(timesGmp), durStd(timesGmp, durMean(timesGmp))

	fmt.Fprintf(os.Stderr, "PrefixSum\tseq=%v±%v\tmanual=%v±%v\tgompher=%v±%v\tspeedup_manual=%.2fx\tspeedup_gompher=%.2fx\tgmp_vs_manual=%.2fx\tcorrect=%v/%v\tphases=%v/%v\n",
		tSeq, tSeqStd, tMan, tManStd, tGmp, tGmpStd,
		float64(tSeq)/float64(tMan), float64(tSeq)/float64(tGmp), float64(tMan)/float64(tGmp),
		rs == rm, rs == rg, sentM == -998, sentG == -998)
}
