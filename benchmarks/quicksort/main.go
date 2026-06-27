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

const (
	N         = 1 << 22
	GrainSize = 1 << 14
)

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

func partition(data []int) int {
	mid := len(data) / 2
	pivot := data[mid]
	data[0], data[mid] = data[mid], data[0]
	j := 0
	for i := 1; i < len(data); i++ {
		if data[i] < pivot {
			j++
			data[i], data[j] = data[j], data[i]
		}
	}
	data[0], data[j] = data[j], data[0]
	return j
}

func quickSortSeq(data []int) {
	if len(data) <= GrainSize {
		sort.Ints(data)
		return
	}
	p := partition(data)
	quickSortSeq(data[:p])
	quickSortSeq(data[p+1:])
}

func quickSortManualRec(data []int, wg *sync.WaitGroup, depth int) {
	defer wg.Done()
	if len(data) <= GrainSize || depth <= 0 {
		sort.Ints(data)
		return
	}
	p := partition(data)
	wg.Add(2)
	go quickSortManualRec(data[:p], wg, depth-1)
	go quickSortManualRec(data[p+1:], wg, depth-1)
}

func quickSortManual(data []int) {
	var wg sync.WaitGroup
	depth := 0
	for n := numProcs(); n > 1; n >>= 1 {
		depth++
	}
	wg.Add(1)
	quickSortManualRec(data, &wg, depth)
	wg.Wait()
}

func quickSortGompher(data []int) {
	if len(data) <= GrainSize {
		sort.Ints(data)
		return
	}
	p := partition(data)
	left, right := data[:p], data[p+1:]
	//gompher task firstprivate(left)
	{
		quickSortGompher(left)
	}
	//gompher task firstprivate(right)
	{
		quickSortGompher(right)
	}
	//gompher taskwait
}

func quickSortGompherEntry(data []int) {
	//gompher taskgroup
	{
		quickSortGompher(data)
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
		quickSortSeq(dS)
		timesSeq[r] = time.Since(t0)
	}
	for r := 0; r < runs; r++ {
		copy(dM, orig)
		t0 := time.Now()
		quickSortManual(dM)
		timesMan[r] = time.Since(t0)
	}
	for r := 0; r < runs; r++ {
		copy(dG, orig)
		t0 := time.Now()
		quickSortGompherEntry(dG)
		timesGmp[r] = time.Since(t0)
	}

	printCSV("quicksort", "seq", timesSeq)
	printCSV("quicksort", "manual", timesMan)
	printCSV("quicksort", "gompher", timesGmp)

	tSeq, tSeqStd := durMean(timesSeq), durStd(timesSeq, durMean(timesSeq))
	tMan, tManStd := durMean(timesMan), durStd(timesMan, durMean(timesMan))
	tGmp, tGmpStd := durMean(timesGmp), durStd(timesGmp, durMean(timesGmp))

	fmt.Fprintf(os.Stderr, "Quicksort\tseq=%v±%v\tmanual=%v±%v\tgompher=%v±%v\tspeedup_manual=%.2fx\tspeedup_gompher=%.2fx\tgmp_vs_manual=%.2fx\tcorrect=%v/%v\n",
		tSeq, tSeqStd, tMan, tManStd, tGmp, tGmpStd,
		float64(tSeq)/float64(tMan), float64(tSeq)/float64(tGmp), float64(tMan)/float64(tGmp),
		isSorted(dM), isSorted(dG))
}
