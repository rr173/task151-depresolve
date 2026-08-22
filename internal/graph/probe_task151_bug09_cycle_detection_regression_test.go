package graph

import (
	"testing"

	"depresolve/internal/model"
)

func TestBug09_CycleDetectionIncludesDependencyLoop(t *testing.T) {
	g := New()
	r := model.Release{ID: "release"}
	g.AddNode("app", r)
	g.AddNode("codec", r)
	g.AddNode("crypto", r)
	g.AddEdge(Edge{From: "app", To: "codec", Satisfied: true})
	g.AddEdge(Edge{From: "codec", To: "crypto", Satisfied: true})
	g.AddEdge(Edge{From: "crypto", To: "app", Satisfied: true})
	cycle, ok := g.CycleFrom("app")
	if !ok || len(cycle) != 4 || cycle[0] != "app" || cycle[len(cycle)-1] != "app" {
		t.Fatalf("cycle=%v found=%v", cycle, ok)
	}
}
