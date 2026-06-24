// Demonstrates data-sharing clauses on a parallel region.
//
// firstprivate(base): each goroutine gets its own copy initialized to the
//                     outer value at the point of the construct.
// private(scratch):   each goroutine gets its own fresh (zero-valued) copy.
//                     The outer scratch is untouched.
// shared(total):      one variable shared by the whole team. Concurrent writes
//                     are synchronized here with atomic.
package main

import "fmt"

func main() {
	base := 100
	total := 0
	scratch := -1

	//gompher parallel firstprivate(base) private(scratch) shared(total)
	{
		scratch = base * 2 // reads the captured base, writes its own copy
		//gompher atomic update
		total += scratch
	}

	fmt.Println("base   (unchanged):", base)
	fmt.Println("scratch(unchanged):", scratch)
	fmt.Println("total  (summed)   :", total)
}
