// Demonstrates data-sharing clauses on a sections construct. Each section runs
// once, on whichever team goroutine claims it.
// - reduction(+:total): each section accumulates into a private copy seeded
//   with the operator identity, and the partials are folded back under a
//   critical section once the sections finish.
// - firstprivate(weight): each section starts from a copy of the captured
//   outer value.
// - lastprivate(stage): the lexically last section's value is written back.
package main

import "fmt"

func main() {
	total := 0
	weight := 10
	stage := -1

	//gompher parallel sections reduction(+:total) firstprivate(weight) lastprivate(stage)
	{
		//gompher section
		{
			total += weight * 1
			stage = 1
		}
		//gompher section
		{
			total += weight * 2
			stage = 2
		}
		//gompher section
		{
			total += weight * 3
			stage = 3
		}
	}

	// total = 10*1 + 10*2 + 10*3 = 60; stage = 3 (the last section).
	fmt.Println("total =", total, "stage =", stage)
}
