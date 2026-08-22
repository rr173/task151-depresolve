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

// An optional dependency that has no satisfiable version must be skipped
// instead of failing the whole release plan. This covers both a plugin whose
// component does not exist and one whose versions do not match the constraint.
func TestSolveSkipsUnsatisfiableOptional(t *testing.T) {
	components := []model.Component{{ID: "app"}, {ID: "lib"}, {ID: "plugin"}}
	releases := []model.Release{
		{ID: "a", ComponentID: "app", Version: "1.0.0", Dependencies: []model.Dependency{
			{TargetComponent: "lib", Constraint: "^1.0.0"},
			{TargetComponent: "plugin", Constraint: "^2.0.0", Optional: true},
			{TargetComponent: "ghost", Constraint: "^1.0.0", Optional: true},
		}},
		{ID: "l1", ComponentID: "lib", Version: "1.0.0"},
		{ID: "p1", ComponentID: "plugin", Version: "1.0.0"},
	}
	c, _ := catalog.New(components, releases)
	out, e := Solve(Input{Request: model.ResolveRequest{RootComponent: "app", RootConstraint: "*", Platform: "linux/amd64"}, Catalog: c})
	if e != nil {
		t.Fatalf("optional unsatisfiable deps must be skipped, got err=%v conflicts=%+v", e, out.Conflicts)
	}
	if len(out.Selections) != 2 {
		t.Fatalf("expected app+lib only, got selections=%+v", out.Selections)
	}
}

// A required dependency with no satisfiable version (missing component or
// non-matching version) must still fail resolution.
func TestSolveFailsUnsatisfiableRequired(t *testing.T) {
	components := []model.Component{{ID: "app"}, {ID: "lib"}, {ID: "plugin"}}
	releases := []model.Release{
		{ID: "a", ComponentID: "app", Version: "1.0.0", Dependencies: []model.Dependency{
			{TargetComponent: "lib", Constraint: "^1.0.0"},
			{TargetComponent: "plugin", Constraint: "^2.0.0"},
			{TargetComponent: "ghost", Constraint: "^1.0.0"},
		}},
		{ID: "l1", ComponentID: "lib", Version: "1.0.0"},
		{ID: "p1", ComponentID: "plugin", Version: "1.0.0"},
	}
	c, _ := catalog.New(components, releases)
	out, e := Solve(Input{Request: model.ResolveRequest{RootComponent: "app", RootConstraint: "*", Platform: "linux/amd64"}, Catalog: c})
	if e == nil {
		t.Fatalf("required unsatisfiable deps must fail, got success selections=%+v", out.Selections)
	}
	if len(out.Conflicts) == 0 {
		t.Fatalf("expected conflicts, got %+v", out)
	}
}
