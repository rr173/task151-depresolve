package catalog

import (
	"testing"

	"depresolve/internal/model"
)

func TestBug06_CaretRangeKeepsCompatibleMinorUpgrade(t *testing.T) {
	c, err := New([]model.Component{{ID: "codec"}}, []model.Release{{ID: "v12", ComponentID: "codec", Version: "1.2.0"}, {ID: "v19", ComponentID: "codec", Version: "1.9.0"}, {ID: "v20", ComponentID: "codec", Version: "2.0.0"}})
	if err != nil {
		t.Fatal(err)
	}
	got, err := c.Candidates("codec", "^1.2.0", "linux/amd64", false)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].ID != "v19" || got[1].ID != "v12" {
		t.Fatalf("caret candidates=%+v", got)
	}
}
