package solver

import (
	"depresolve/internal/catalog"
	"depresolve/internal/model"
	"testing"
)

func TestSolveChoosesHighestCompatible(t *testing.T) {
	c, _ := catalog.New([]model.Component{{ID: "app"}, {ID: "lib"}}, []model.Release{{ID: "a", ComponentID: "app", Version: "1.0.0", Dependencies: []model.Dependency{{TargetComponent: "lib", Constraint: "^1.0.0"}}}, {ID: "l1", ComponentID: "lib", Version: "1.0.0"}, {ID: "l2", ComponentID: "lib", Version: "1.2.0"}})
	out, e := Solve(Input{Request: model.ResolveRequest{RootComponent: "app", RootConstraint: "*", Platform: "linux/amd64"}, Catalog: c})
	if e != nil {
		t.Fatal(e)
	}
	if len(out.Selections) != 2 || out.Selections[1].Version != "1.2.0" {
		t.Fatalf("selections=%+v", out.Selections)
	}
}

func TestSolveRejectsConflictingCapability(t *testing.T) {
	// lib provides capability "http2"; codec declares it as a conflict.
	// Both are pulled in by app -> the combination is mutually exclusive.
	c, _ := catalog.New(
		[]model.Component{{ID: "app"}, {ID: "lib"}, {ID: "codec"}},
		[]model.Release{
			{ID: "app-1", ComponentID: "app", Version: "1.0.0", Dependencies: []model.Dependency{
				{TargetComponent: "lib", Constraint: "^1.0.0"},
				{TargetComponent: "codec", Constraint: "^1.0.0"},
			}},
			{ID: "lib-1", ComponentID: "lib", Version: "1.0.0", Capabilities: []string{"http2"}},
			{ID: "codec-1", ComponentID: "codec", Version: "1.0.0", Conflicts: []string{"http2"}},
		},
	)
	out, err := Solve(Input{Request: model.ResolveRequest{RootComponent: "app", RootConstraint: "*", Platform: "any"}, Catalog: c})
	if err == nil {
		t.Fatalf("expected capability conflict, got selections=%+v", out.Selections)
	}
	var found bool
	for _, c := range out.Conflicts {
		if c.Kind == model.ConflictCapability {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected a %q conflict, got %+v", model.ConflictCapability, out.Conflicts)
	}
}

func TestSolveRejectsSelfConflictingCapability(t *testing.T) {
	// A release that both provides and conflicts with the same capability is rejected.
	c, _ := catalog.New(
		[]model.Component{{ID: "app"}, {ID: "lib"}},
		[]model.Release{
			{ID: "app-1", ComponentID: "app", Version: "1.0.0", Dependencies: []model.Dependency{{TargetComponent: "lib", Constraint: "^1.0.0"}}},
			{ID: "lib-1", ComponentID: "lib", Version: "1.0.0", Capabilities: []string{"x"}, Conflicts: []string{"x"}},
		},
	)
	_, err := Solve(Input{Request: model.ResolveRequest{RootComponent: "app", RootConstraint: "*", Platform: "any"}, Catalog: c})
	if err == nil {
		t.Fatal("expected self-conflict to be rejected")
	}
}

func TestSolveAllowsCompatibleCapabilities(t *testing.T) {
	// Two releases providing the same capability is fine; a conflict only arises
	// when one explicitly declares the other's capability as incompatible.
	c, _ := catalog.New(
		[]model.Component{{ID: "app"}, {ID: "lib"}, {ID: "util"}},
		[]model.Release{
			{ID: "app-1", ComponentID: "app", Version: "1.0.0", Dependencies: []model.Dependency{
				{TargetComponent: "lib", Constraint: "^1.0.0"},
				{TargetComponent: "util", Constraint: "^1.0.0"},
			}},
			{ID: "lib-1", ComponentID: "lib", Version: "1.0.0", Capabilities: []string{"logging"}},
			{ID: "util-1", ComponentID: "util", Version: "1.0.0", Capabilities: []string{"logging"}},
		},
	)
	out, err := Solve(Input{Request: model.ResolveRequest{RootComponent: "app", RootConstraint: "*", Platform: "any"}, Catalog: c})
	if err != nil {
		t.Fatalf("compatible capabilities must resolve: %v", err)
	}
	if len(out.Selections) != 3 {
		t.Fatalf("selections=%+v", out.Selections)
	}
}
