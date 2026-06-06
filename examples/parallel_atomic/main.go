package main

import "fmt"

func main() {
	var counter int
	var flag int

	//gompher parallel
	{
		//gompher atomic update
		counter++

		//gompher atomic write
		flag = 1
	}

	var snapshot int
	//gompher atomic read
	snapshot = flag

	fmt.Println("counter:", counter)
	fmt.Println("flag:", snapshot)
}
