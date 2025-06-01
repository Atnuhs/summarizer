package main

import (
	"fmt"
	"sort"
)

func main() {
	numbers := []int{3, 1, 4, 1, 5, 9, 2, 6}
	sort.Ints(numbers)
	
	for i, num := range numbers {
		if i > 0 {
			fmt.Print(" ")
		}
		fmt.Print(num)
	}
	fmt.Println()
}