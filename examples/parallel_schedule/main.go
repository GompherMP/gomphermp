package main

import (
	"fmt"
	"math"
)

func heavyWork(i int) float64 {
	result := 0.0
	for k := 0; k < (i+1)*1000; k++ {
		result += math.Sqrt(float64(k))
	}
	return result
}

func main() {
	const N = 20
	results := make([]float64, N)

	//gompher parallel for schedule(dynamic, 4)
	for i := 0; i < N; i++ {
		results[i] = heavyWork(i)
	}

	fmt.Printf("result[0]=%.2f  result[19]=%.2f\n", results[0], results[19])
}
