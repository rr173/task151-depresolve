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
