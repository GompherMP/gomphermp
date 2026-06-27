package main

import (
	"fmt"
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

func main() {
	orig := newData()
	runs := 3
	dS, dM, dG := make([]int, N), make([]int, N), make([]int, N)

	t0 := time.Now()
	for r := 0; r < runs; r++ {
		copy(dS, orig)
		quickSortSeq(dS)
	}
	tSeq := time.Since(t0) / time.Duration(runs)

	t0 = time.Now()
	for r := 0; r < runs; r++ {
		copy(dM, orig)
		quickSortManual(dM)
	}
	tMan := time.Since(t0) / time.Duration(runs)

	t0 = time.Now()
	for r := 0; r < runs; r++ {
		copy(dG, orig)
		quickSortGompherEntry(dG)
	}
	tGmp := time.Since(t0) / time.Duration(runs)

	fmt.Printf("Quicksort\tseq=%v\tmanual=%v\tgompher=%v\tspeedup_manual=%.2fx\tspeedup_gompher=%.2fx\tgmp_vs_manual=%.2fx\tcorrect=%v/%v\n",
		tSeq, tMan, tGmp,
		float64(tSeq)/float64(tMan), float64(tSeq)/float64(tGmp), float64(tMan)/float64(tGmp),
		isSorted(dM), isSorted(dG))
}
