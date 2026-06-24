package main

import (
	"fmt"
	"time"
)

func main() {
	//gompher taskgroup
	{
		//gompher task
		{
			fmt.Println("task 1: started")
			time.Sleep(30 * time.Millisecond)
			fmt.Println("task 1: done")
		}

		//gompher task
		{
			fmt.Println("task 2: started")
			time.Sleep(20 * time.Millisecond)
			fmt.Println("task 2: done")
		}

		//gompher task
		{
			fmt.Println("task 3: started")
			time.Sleep(10 * time.Millisecond)
			fmt.Println("task 3: done")
		}
	}

	fmt.Println("all tasks finished")
}
