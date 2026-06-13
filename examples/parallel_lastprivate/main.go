// Demonstrates the lastprivate clause: each goroutine works on a private copy
// of the variable during the loop, but the value produced by the
// sequentially-last iteration (i == N-1) is copied back out, exactly as if the
// loop had run serially.
package main

import "fmt"

func main() {
	const N = 10

	last := -1
	//gompher parallel for lastprivate(last)
	for i := 0; i < N; i++ {
		last = i * i
	}

	// The last iteration is i == 9, so last holds 9*9 = 81, the same value a
	// serial loop would have left behind.
	fmt.Println("last square =", last) // 81
}
