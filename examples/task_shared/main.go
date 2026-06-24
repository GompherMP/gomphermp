package main

import "fmt"

func main() {
	var counter int

	//gompher taskgroup
	{
		for i := 0; i < 5; i++ {
			//gompher task shared(counter)
			{
				//gompher critical
				{
					counter++
				}
			}
		}
	}

	fmt.Println("final counter:", counter)
}
