// Demonstrates the reduction clause: each goroutine accumulates a private
// partial result, and the partials are combined into the shared variable once
// the loop finishes.
package main

import "fmt"

func main() {
	const N = 1000

	sum := 0
	//gompher parallel for reduction(+:sum)
	for i := 0; i < N; i++ {
		sum += i
	}

	fmt.Println("sum of 0..999 =", sum) // 499500
}
