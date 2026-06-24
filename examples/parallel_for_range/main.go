// Demonstrates GompherMP's support for the full OpenMP canonical loop form, not
// just `for i := 0; i < N; i++`.
package main

import "fmt"

func main() {
	const N = 100

	sum := 0
	//gompher parallel for reduction(+:sum)
	for i := 1; i <= N; i++ {
		sum += i
	}

	// 1 + 2 + ... + 100 = 5050.
	fmt.Println("sum of 1..100 =", sum)
}
