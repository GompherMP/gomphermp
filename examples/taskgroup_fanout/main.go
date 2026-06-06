package main

import "fmt"

func doWork(id int) {
	fmt.Printf("work %d done\n", id)
}

func main() {
	//gompher taskgroup
	{
		//gompher task
		{
			//gompher task
			{
				doWork(1)
			}
			//gompher task
			{
				doWork(2)
			}
		}

		//gompher task
		{
			doWork(3)
		}
	}

	fmt.Println("all work complete, including nested tasks")
}
