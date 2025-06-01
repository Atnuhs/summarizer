package main

import (
	"fmt"
	"with-deps/graph"
	"with-deps/unionfind"
)

func main() {
	uf := unionfind.New(5)
	uf.Union(0, 1)
	uf.Union(2, 3)
	
	fmt.Printf("0 and 1 connected: %t\n", uf.Connected(0, 1))
	fmt.Printf("0 and 2 connected: %t\n", uf.Connected(0, 2))
	
	g := graph.New(4)
	g.AddEdge(0, 1)
	g.AddEdge(1, 2)
	g.AddEdge(2, 3)
	
	fmt.Println("Graph structure:")
	g.PrintGraph()
	
	fmt.Print("DFS traversal from 0: ")
	visited := make([]bool, 4)
	g.DFS(0, visited)
	fmt.Println()
}