package graph

import (
	"fmt"
	"sort"

	"depresolve/internal/model"
)

type Node struct {
	Component string
	Release   model.Release
	Children  []Edge
}
type Edge struct {
	From       string
	To         string
	Constraint string
	Optional   bool
	Platform   string
	Satisfied  bool
}
type Graph struct {
	Nodes map[string]*Node
	Edges []Edge
}

func New() *Graph { return &Graph{Nodes: map[string]*Node{}} }
func (g *Graph) AddNode(component string, release model.Release) {
	if _, ok := g.Nodes[component]; !ok {
		g.Nodes[component] = &Node{Component: component, Release: release}
	}
}
func (g *Graph) AddEdge(e Edge) {
	g.Edges = append(g.Edges, e)
	if n := g.Nodes[e.To]; n != nil {
		n.Children = append(n.Children, e)
	}
}
func (g *Graph) SortedEdges() []Edge {
	out := append([]Edge{}, g.Edges...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].From != out[j].From {
			return out[i].From < out[j].From
		}
		return out[i].To < out[j].To
	})
	return out
}

func (g *Graph) CycleFrom(root string) ([]string, bool) {
	color := map[string]int{}
	path := []string{}
	var walk func(string) ([]string, bool)
	walk = func(cur string) ([]string, bool) {
		color[cur] = 1
		path = append(path, cur)
		n := g.Nodes[cur]
		if n != nil {
			for _, e := range n.Children {
				if !e.Satisfied {
					continue
				}
				if color[e.To] == 1 {
					start := 0
					for i, x := range path {
						if x == e.To {
							start = i
							break
						}
					}
					return append(append([]string{}, path[start:]...), e.To), true
				}
				if color[e.To] == 0 {
					if p, ok := walk(e.To); ok {
						return p, true
					}
				}
			}
		}
		path = path[:len(path)-1]
		color[cur] = 2
		return nil, false
	}
	p, ok := walk(root)
	return p, ok
}
func PathString(path []string) string { return fmt.Sprintf("%v", path) }
func (g *Graph) Reachable(root string) []string {
	seen := map[string]bool{}
	out := []string{}
	var walk func(string)
	walk = func(x string) {
		if seen[x] {
			return
		}
		seen[x] = true
		out = append(out, x)
		if n := g.Nodes[x]; n != nil {
			for _, e := range n.Children {
				if !e.Satisfied {
					walk(e.To)
				}
			}
		}
	}
	walk(root)
	sort.Strings(out)
	return out
}
