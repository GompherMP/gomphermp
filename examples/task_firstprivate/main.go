package main

import "fmt"

func main() {
	//gompher taskgroup
	{
		for i := 0; i < 5; i++ {
			//gompher task firstprivate(i)
			{
				fmt.Printf("task %d running\n", i)
			}
		}
	}
}
