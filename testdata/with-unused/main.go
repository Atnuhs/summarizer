package main

import (
	"fmt"
	"with-unused/math"
	"with-unused/utils"
)

func main() {
	// Only use specific functions from math package
	result := math.Add(5, 3)
	product := math.Multiply(result, 2)
	
	fmt.Printf("5 + 3 = %d\n", result)
	fmt.Printf("Result * 2 = %d\n", product)
	
	// Use Calculator type and its methods
	calc := &math.Calculator{}
	calc.Add(10)
	calc.Add(5)
	fmt.Printf("Calculator result: %d\n", calc.GetResult())
	
	// Only use specific functions from utils package
	utils.PrintMessage("Hello, World!")
	
	logger := utils.NewLogger("TEST")
	logger.Log("This is a test message")
	
	// Note: We do NOT use:
	// - math.Subtract, math.Divide
	// - math.UnusedStruct, math.AnotherUnusedStruct
	// - math.NewUnusedStruct, math.UnusedGlobalFunction
	// - math.UnusedConstant, math.UnusedVariable
	// - utils.FormatNumber, utils.ReverseString
	// - utils.FileManager
	// - utils.DefaultPath, utils.GlobalCounter
}