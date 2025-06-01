package main

import (
	"fmt"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

func main() {
	// Using samber/lo for functional programming utilities
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	
	// Filter even numbers
	evens := lo.Filter(numbers, func(n int, _ int) bool {
		return n%2 == 0
	})
	
	// Map to squares
	squares := lo.Map(evens, func(n int, _ int) int {
		return n * n
	})
	
	fmt.Printf("Even numbers: %v\n", evens)
	fmt.Printf("Squares: %v\n", squares)
	
	// Using samber/mo for optional values
	opt1 := mo.Some(42)
	opt2 := mo.None[int]()
	
	if val, ok := opt1.Get(); ok {
		fmt.Printf("opt1 has value: %d\n", val)
	}
	
	if val, ok := opt2.Get(); ok {
		fmt.Printf("opt2 has value: %d\n", val)
	} else {
		fmt.Println("opt2 is empty")
	}
	
	// Chain operations
	result := opt1.Map(func(x int) (int, bool) {
		return x * 2, true
	}).OrElse(0)
	
	fmt.Printf("Result: %d\n", result)
}