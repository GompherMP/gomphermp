package main

import "fmt"

const teamSize = 4

var localData [teamSize]int

func initPhase(id int) {
	localData[id] = id * 10
	fmt.Printf("goroutine %d initialized data: %d\n", id, localData[id])
}

func computePhase(id int) {
	sum := 0
	for _, v := range localData {
		sum += v
	}
	fmt.Printf("goroutine %d sees total: %d\n", id, sum)
}

func main() {
	//gompher parallel
	{
		initPhase(0)

		//gompher barrier

		computePhase(0)
	}
}
