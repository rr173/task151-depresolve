package plan

import (
	"depresolve/internal/model"
	"testing"
)

func TestDiff(t *testing.T) {
	a := []model.Selection{{ComponentID: "a", Version: "1.0.0"}, {ComponentID: "b", Version: "1.0.0"}}
	b := []model.Selection{{ComponentID: "a", Version: "2.0.0"}, {ComponentID: "c", Version: "1.0.0"}}
	d := Diff(a, b)
	if len(d.Changed) != 1 || len(d.Added) != 1 || len(d.Removed) != 1 {
		t.Fatalf("diff=%+v", d)
	}
}
