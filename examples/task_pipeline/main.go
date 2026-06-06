package main

import "fmt"

func main() {
	var data int

	//gompher taskgroup
	{
		//gompher task depend(out:data)
		{
			data = 42
		}

		//gompher task depend(inout:data)
		{
			data = data * 2
		}

		//gompher task depend(in:data)
		{
			fmt.Println("result:", data)
		}
	}
}
