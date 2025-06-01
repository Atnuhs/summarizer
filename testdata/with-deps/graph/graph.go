package graph

import "fmt"

type Graph struct {
	adj [][]int
	n   int
}

func New(n int) *Graph {
	return &Graph{
		adj: make([][]int, n),
		n:   n,
	}
}

func (g *Graph) AddEdge(u, v int) {
	g.adj[u] = append(g.adj[u], v)
	g.adj[v] = append(g.adj[v], u)
}

func (g *Graph) PrintGraph() {
	for i := 0; i < g.n; i++ {
		fmt.Printf("Vertex %d: ", i)
		for _, v := range g.adj[i] {
			fmt.Printf("%d ", v)
		}
		fmt.Println()
	}
}

func (g *Graph) DFS(start int, visited []bool) {
	visited[start] = true
	fmt.Printf("%d ", start)
	
	for _, v := range g.adj[start] {
		if !visited[v] {
			g.DFS(v, visited)
		}
	}
}