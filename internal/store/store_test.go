package store

import (
	"context"
	"depresolve/internal/model"
	"path/filepath"
	"testing"
	"time"
)

func TestRestartKeepsCatalog(t *testing.T) {
	p := filepath.Join(t.TempDir(), "x.db")
	ctx := context.Background()
	s, e := Open(p)
	if e != nil {
		t.Fatal(e)
	}
	now := time.Now()
	if e = s.CreateComponent(ctx, model.Component{ID: "a", Name: "A", Ecosystem: "go", StableOnly: true, CreatedAt: now, UpdatedAt: now}); e != nil {
		t.Fatal(e)
	}
	s.Close()
	s, e = Open(p)
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	xs, e := s.ListComponents(ctx)
	if e != nil || len(xs) != 1 || xs[0].ID != "a" {
		t.Fatalf("components=%+v err=%v", xs, e)
	}
}

func TestReleaseCapabilitiesConflictsRoundTrip(t *testing.T) {
	p := filepath.Join(t.TempDir(), "x.db")
	ctx := context.Background()
	s, e := Open(p)
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	now := time.Now()
	if e = s.CreateComponent(ctx, model.Component{ID: "a", Name: "A", Ecosystem: "go", CreatedAt: now, UpdatedAt: now}); e != nil {
		t.Fatal(e)
	}
	r := model.Release{ID: "r1", ComponentID: "a", Version: "1.0.0", Capabilities: []string{"http2"}, Conflicts: []string{"http1"}}
	if e = s.CreateRelease(ctx, r); e != nil {
		t.Fatal(e)
	}
	got, e := s.ListReleases(ctx, "a")
	if e != nil {
		t.Fatal(e)
	}
	if len(got) != 1 {
		t.Fatalf("got=%+v", got)
	}
	if !eqStr(got[0].Capabilities, r.Capabilities) || !eqStr(got[0].Conflicts, r.Conflicts) {
		t.Fatalf("ListReleases round-trip mismatch: got cap=%v conf=%v, want cap=%v conf=%v", got[0].Capabilities, got[0].Conflicts, r.Capabilities, r.Conflicts)
	}
	all, e := s.ListAllReleases(ctx)
	if e != nil {
		t.Fatal(e)
	}
	if len(all) != 1 || !eqStr(all[0].Capabilities, r.Capabilities) || !eqStr(all[0].Conflicts, r.Conflicts) {
		t.Fatalf("ListAllReleases round-trip mismatch: got cap=%v conf=%v", all[0].Capabilities, all[0].Conflicts)
	}
	one, e := s.GetRelease(ctx, "r1")
	if e != nil {
		t.Fatal(e)
	}
	if !eqStr(one.Capabilities, r.Capabilities) || !eqStr(one.Conflicts, r.Conflicts) {
		t.Fatalf("GetRelease mismatch: got cap=%v conf=%v, want cap=%v conf=%v", one.Capabilities, one.Conflicts, r.Capabilities, r.Conflicts)
	}
}

func eqStr(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
