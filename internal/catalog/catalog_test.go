package catalog

import (
	"depresolve/internal/model"
	"testing"
)

func TestCandidatesPreferStable(t *testing.T) {
	c, e := New([]model.Component{{ID: "a"}}, []model.Release{{ID: "r1", ComponentID: "a", Version: "1.0.0"}, {ID: "r2", ComponentID: "a", Version: "1.1.0-alpha.1"}})
	if e != nil {
		t.Fatal(e)
	}
	got, e := c.Candidates("a", "*", "", false)
	if e != nil || len(got) != 1 || got[0].ID != "r1" {
		t.Fatalf("got=%+v err=%v", got, e)
	}
}
