package main

import (
	"fmt"
	"math"
	"os"
	goruntime "runtime"
	"sort"
	"sync"
	"time"
)

const N = 1 << 20
const GrainSize = 1 << 14

func numProcs() int { return goruntime.GOMAXPROCS(0) }

func isSorted(d []int) bool {
	for i := 1; i < len(d); i++ {
		if d[i] < d[i-1] {
			return false
		}
	}
	return true
}
func newData() []int {
	d := make([]int, N)
	for i := range d {
		d[i] = N - i
	}
	return d
}
func mergeParts(data []int, mid int) {
	tmp := make([]int, len(data))
	i, j, k := 0, mid, 0
	for i < mid && j < len(data) {
		if data[i] <= data[j] {
			tmp[k] = data[i]
			i++
		} else {
			tmp[k] = data[j]
			j++
		}
		k++
	}
	for i < mid {
		tmp[k] = data[i]
		i++
		k++
	}
	for j < len(data) {
		tmp[k] = data[j]
		j++
		k++
	}
	copy(data, tmp)
}

func mergeSortSeq(data []int) {
	if len(data) <= GrainSize {
		sort.Ints(data)
		return
	}
	mid := len(data) / 2
	mergeSortSeq(data[:mid])
	mergeSortSeq(data[mid:])
	mergeParts(data, mid)
}

func mergeSortManual(data []int) {
	p, blockSize := numProcs(), len(data)/numProcs()
	var wg sync.WaitGroup
	for t := 0; t < p; t++ {
		s, e := t*blockSize, (t+1)*blockSize
		if t == p-1 {
			e = len(data)
		}
		wg.Add(1)
		go func(s, e int) { defer wg.Done(); sort.Ints(data[s:e]) }(s, e)
	}
	wg.Wait()
	for b := 1; b < p; b++ {
		mid := b * blockSize
		if mid > len(data) {
			break
		}
		hi := (b + 1) * blockSize
		if b == p-1 || hi > len(data) {
			hi = len(data)
		}
		mergeParts(data[:hi], mid)
	}
}

func mergeSortGompher(data []int) {
	p := numProcs()
	blockSize := len(data) / p
	//gompher taskgroup
	{
		//gompher taskloop grainsize(1)
		for b := 0; b < p; b++ {
			s, e := b*blockSize, (b+1)*blockSize
			if b == p-1 {
				e = len(data)
			}
			sort.Ints(data[s:e])
		}
	}
	for b := 1; b < p; b++ {
		mid := b * blockSize
		if mid > len(data) {
			break
		}
		hi := (b + 1) * blockSize
		if b == p-1 || hi > len(data) {
			hi = len(data)
		}
		mergeParts(data[:hi], mid)
	}
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
	orig := newData()
	const runs = 10
	dS, dM, dG := make([]int, N), make([]int, N), make([]int, N)
	timesSeq := make([]time.Duration, runs)
	timesMan := make([]time.Duration, runs)
	timesGmp := make([]time.Duration, runs)

	for r := 0; r < runs; r++ {
		copy(dS, orig)
		t0 := time.Now()
		mergeSortSeq(dS)
		timesSeq[r] = time.Since(t0)
	}
	for r := 0; r < runs; r++ {
		copy(dM, orig)
		t0 := time.Now()
		mergeSortManual(dM)
		timesMan[r] = time.Since(t0)
	}
	for r := 0; r < runs; r++ {
		copy(dG, orig)
		t0 := time.Now()
		mergeSortGompher(dG)
		timesGmp[r] = time.Since(t0)
	}

	printCSV("mergesort", "seq", timesSeq)
	printCSV("mergesort", "manual", timesMan)
	printCSV("mergesort", "gompher", timesGmp)

	tSeq, tSeqStd := durMean(timesSeq), durStd(timesSeq, durMean(timesSeq))
	tMan, tManStd := durMean(timesMan), durStd(timesMan, durMean(timesMan))
	tGmp, tGmpStd := durMean(timesGmp), durStd(timesGmp, durMean(timesGmp))

	fmt.Fprintf(os.Stderr, "Mergesort\tseq=%v±%v\tmanual=%v±%v\tgompher=%v±%v\tspeedup_manual=%.2fx\tspeedup_gompher=%.2fx\tgmp_vs_manual=%.2fx\tcorrect=%v/%v\n",
		tSeq, tSeqStd, tMan, tManStd, tGmp, tGmpStd,
		float64(tSeq)/float64(tMan), float64(tSeq)/float64(tGmp), float64(tMan)/float64(tGmp),
		isSorted(dM), isSorted(dG))
}
