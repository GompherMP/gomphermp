// Demonstrates the worksharing model: a single parallel region whose team is
// reused across several phases, each closed by an implicit barrier.
//
// Phase 1 (for): the team splits the loop and fills the array.
// Phase 2 (single): exactly one goroutine reports the checkpoint.
// Phase 3 (for): the same team splits the loop again and doubles each value.
package main

import "fmt"

func main() {
	const N = 100
	data := make([]int, N)

	//gompher parallel
	{
		//gompher for
		for i := 0; i < N; i++ {
			data[i] = i
		}

		//gompher single
		{
			fmt.Println("phase 1 complete: array filled")
		}

		//gompher for
		for i := 0; i < N; i++ {
			data[i] *= 2
		}
	}

	sum := 0
	for _, v := range data {
		sum += v
	}
	fmt.Println("sum after doubling:", sum)
}
