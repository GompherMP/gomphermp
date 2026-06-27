package main

import (
	"fmt"
	"math"
	goruntime "runtime"
	"time"
)

const (
	ChunkCount = 4
	ChunkSize  = 1 << 20
)

func numProcs() int { return goruntime.GOMAXPROCS(0) }

func stage1(input []float64, output []float64) {
	for i, v := range input {
		output[i] = math.Sqrt(v)
	}
}

func stage2(data []float64) float64 {
	s := 0.0
	for _, v := range data {
		s += math.Log1p(v)
	}
	return s
}

func pipelineSeq(chunks [][]float64, bufs [][]float64) float64 {
	total := 0.0
	for i := range chunks {
		stage1(chunks[i], bufs[i])
		total += stage2(bufs[i])
	}
	return total
}

func pipelineManual(chunks [][]float64, bufs [][]float64) float64 {
	p := numProcs()
	results := make([]float64, ChunkCount)
	ch := make(chan int, ChunkCount)
	for i := range chunks {
		ch <- i
	}
	close(ch)
	done := make(chan struct{}, p)
	for w := 0; w < p; w++ {
		go func() {
			for i := range ch {
				stage1(chunks[i], bufs[i])
				results[i] = stage2(bufs[i])
			}
			done <- struct{}{}
		}()
	}
	for w := 0; w < p; w++ {
		<-done
	}
	total := 0.0
	for _, v := range results {
		total += v
	}
	return total
}

func pipelineGompher(chunks [][]float64, bufs [][]float64) float64 {
	// s0..s3: sentinels set after stage1 completes for each chunk.
	// f0..f3: stage2 results per chunk.
	// Sentinels decouple the array write from the dependency token so that
	// depend() only needs to name simple scalar variables.
	var s0, s1, s2, s3 float64
	var f0, f1, f2, f3 float64
	var total float64

	//gompher taskgroup
	{
		//gompher task depend(out:s0)
		{
			stage1(chunks[0], bufs[0])
			s0 = 1
		}
		//gompher task depend(out:s1)
		{
			stage1(chunks[1], bufs[1])
			s1 = 1
		}
		//gompher task depend(out:s2)
		{
			stage1(chunks[2], bufs[2])
			s2 = 1
		}
		//gompher task depend(out:s3)
		{
			stage1(chunks[3], bufs[3])
			s3 = 1
		}
		//gompher task depend(in:s0) depend(out:f0)
		{
			f0 = stage2(bufs[0])
		}
		//gompher task depend(in:s1) depend(out:f1)
		{
			f1 = stage2(bufs[1])
		}
		//gompher task depend(in:s2) depend(out:f2)
		{
			f2 = stage2(bufs[2])
		}
		//gompher task depend(in:s3) depend(out:f3)
		{
			f3 = stage2(bufs[3])
		}
		//gompher task depend(in:f0, f1, f2, f3) depend(out:total)
		{
			total = f0 + f1 + f2 + f3
		}
	}
	return total
}

func main() {
	chunks := make([][]float64, ChunkCount)
	bufs := make([][]float64, ChunkCount)
	for i := range chunks {
		chunks[i] = make([]float64, ChunkSize)
		bufs[i] = make([]float64, ChunkSize)
		for j := range chunks[i] {
			chunks[i][j] = float64(j+1) + float64(i)*0.01
		}
	}
	runs := 5

	t0 := time.Now()
	var rs float64
	for r := 0; r < runs; r++ {
		rs = pipelineSeq(chunks, bufs)
	}
	tSeq := time.Since(t0) / time.Duration(runs)

	t0 = time.Now()
	var rm float64
	for r := 0; r < runs; r++ {
		rm = pipelineManual(chunks, bufs)
	}
	tMan := time.Since(t0) / time.Duration(runs)

	t0 = time.Now()
	var rg float64
	for r := 0; r < runs; r++ {
		rg = pipelineGompher(chunks, bufs)
	}
	tGmp := time.Since(t0) / time.Duration(runs)

	eps := math.Abs(rs) * 1e-9
	fmt.Printf("Pipeline\tseq=%v\tmanual=%v\tgompher=%v\tspeedup_manual=%.2fx\tspeedup_gompher=%.2fx\tgmp_vs_manual=%.2fx\tcorrect=%v/%v\n",
		tSeq, tMan, tGmp,
		float64(tSeq)/float64(tMan), float64(tSeq)/float64(tGmp), float64(tMan)/float64(tGmp),
		math.Abs(rm-rs) < eps, math.Abs(rg-rs) < eps)
}
