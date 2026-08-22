package catalog

import (
	"testing"

	"depresolve/internal/model"
)

func TestBug03_StableResolutionExcludesPrerelease(t *testing.T) {
	c, err := New([]model.Component{{ID: "codec"}}, []model.Release{{ID: "stable", ComponentID: "codec", Version: "1.2.0"}, {ID: "candidate", ComponentID: "codec", Version: "1.3.0-rc.1"}})
	if err != nil {
		t.Fatal(err)
	}
	got, err := c.Candidates("codec", "*", "linux/amd64", false)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != "stable" {
		t.Fatalf("stable resolution admitted prerelease: %+v", got)
	}
}
