package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	var inputFile = flag.String("input", "", "Input Go file path (required)")
	var outputFile = flag.String("output", "submit.go", "Output file path")
	flag.Parse()

	if *inputFile == "" {
		fmt.Fprintf(os.Stderr, "Error: -input flag is required\n")
		flag.Usage()
		os.Exit(1)
	}

	fmt.Printf("Analyzing %s...\n", *inputFile)

	bundler := NewBundler()
	if err := bundler.Bundle(*inputFile, *outputFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Done!")
}