package graph

import (
	"reflect"
	"testing"

	"depresolve/internal/model"
)

func TestBug05_ReachableIncludesTransitiveDependencies(t *testing.T) {
	g := New()
	r := model.Release{ID: "release"}
	for _, id := range []string{"app", "codec", "crypto"} {
		g.AddNode(id, r)
	}
	g.AddEdge(Edge{From: "app", To: "codec", Satisfied: true})
	g.AddEdge(Edge{From: "codec", To: "crypto", Satisfied: true})
	if got, want := g.Reachable("app"), []string{"app", "codec", "crypto"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("reachable=%v want=%v", got, want)
	}
}
