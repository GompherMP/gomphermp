package main

import "fmt"

func main() {
	const N = 16
	results := make([]int, N)

	//gompher parallel for
	for i := 0; i < N; i++ {
		results[i] = i * i
	}

	sum := 0
	for _, v := range results {
		sum += v
	}
	fmt.Println("sum of squares 0..15:", sum) // 1240
}
