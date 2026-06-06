package main

import "fmt"

const N = 16

func main() {
	results := make([]int, N)

	//gompher taskgroup
	{
		//gompher taskloop grainsize(4)
		for i := 0; i < N; i++ {
			results[i] = i * i
		}
	}

	sum := 0
	for _, v := range results {
		sum += v
	}
	fmt.Println("sum of squares 0..15:", sum)
}
