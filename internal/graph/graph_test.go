package graph

import (
	"depresolve/internal/model"
	"testing"
)

func TestCycle(t *testing.T) {
	g := New()
	r := model.Release{ID: "r"}
	g.AddNode("a", r)
	g.AddNode("b", r)
	g.AddEdge(Edge{From: "a", To: "b", Satisfied: true})
	g.AddEdge(Edge{From: "b", To: "a", Satisfied: true})
	if p, ok := g.CycleFrom("a"); !ok || len(p) != 3 {
		t.Fatalf("cycle=%v ok=%v", p, ok)
	}
}
