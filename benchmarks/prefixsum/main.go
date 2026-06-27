package main

import (
	"fmt"
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

func main() {
	data := make([]float64, N)
	for i := range data {
		data[i] = 1.0
	}
	runs := 5

	t0 := time.Now()
	var rs float64
	for r := 0; r < runs; r++ {
		rs = sumSeq(data)
	}
	tSeq := time.Since(t0) / time.Duration(runs)

	t0 = time.Now()
	var rm float64
	for r := 0; r < runs; r++ {
		rm = sumManual(data)
	}
	tMan := time.Since(t0) / time.Duration(runs)

	t0 = time.Now()
	var rg float64
	for r := 0; r < runs; r++ {
		rg = sumGompher(data)
	}
	tGmp := time.Since(t0) / time.Duration(runs)

	small := make([]int, 100)
	for i := range small {
		small[i] = i
	}
	sentM := phasesManual(small)
	for i := range small {
		small[i] = i
	}
	sentG := phasesGompher(small)

	fmt.Printf("PrefixSum\tseq=%v\tmanual=%v\tgompher=%v\tspeedup_manual=%.2fx\tspeedup_gompher=%.2fx\tgmp_vs_manual=%.2fx\tcorrect=%v/%v\tphases=%v/%v\n",
		tSeq, tMan, tGmp,
		float64(tSeq)/float64(tMan), float64(tSeq)/float64(tGmp), float64(tMan)/float64(tGmp),
		rs == rm, rs == rg, sentM == -998, sentG == -998)
}
