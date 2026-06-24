package main

import "fmt"

func calculate(i int) int {
	return i * i
}

func main() {
	var result int = -1

	//gompher taskgroup
	{
		for i := 0; i < 4; i++ {
			//gompher task private(result)
			{
				result = calculate(i)
				fmt.Printf("task %d local result: %d\n", i, result)
			}
		}
	}

	fmt.Println("outer result unchanged:", result)
}
