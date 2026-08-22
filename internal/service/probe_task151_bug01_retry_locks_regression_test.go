package service

import (
	"context"
	"path/filepath"
	"testing"

	"depresolve/internal/model"
	"depresolve/internal/store"
)

func TestBug01_RetryPreservesLockedVersions(t *testing.T) {
	ctx := context.Background()
	st, err := store.Open(filepath.Join(t.TempDir(), "resolver.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	svc, err := New(st)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []model.CreateComponentRequest{{ID: "app", Name: "App", Ecosystem: "go"}, {ID: "lib", Name: "Library", Ecosystem: "go"}} {
		if _, err := svc.CreateComponent(ctx, c); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := svc.CreateRelease(ctx, "app", model.CreateReleaseRequest{ID: "app-1", Version: "1.0.0", Dependencies: []model.DependencyInput{{ID: "app-lib", TargetComponent: "lib", Constraint: "^1.0.0"}}}); err != nil {
		t.Fatal(err)
	}
	for _, release := range []model.CreateReleaseRequest{{ID: "lib-1", Version: "1.0.0"}, {ID: "lib-2", Version: "1.9.0"}} {
		if _, err := svc.CreateRelease(ctx, "lib", release); err != nil {
			t.Fatal(err)
		}
	}
	first, err := svc.Resolve(ctx, model.ResolveRequest{ID: "locked", RootComponent: "app", RootConstraint: "*", Locks: map[string]string{"lib": "1.0.0"}, Policy: "stable"})
	if err != nil {
		t.Fatal(err)
	}
	if got := selectionVersion(first.Selections, "lib"); got != "1.0.0" {
		t.Fatalf("first resolve chose %q", got)
	}
	retry, err := svc.RetryResolve(ctx, first.Run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got := selectionVersion(retry.Selections, "lib"); got != "1.0.0" {
		t.Fatalf("retry lost the locked lib version: got %q", got)
	}
}

func selectionVersion(items []model.Selection, component string) string {
	for _, item := range items {
		if item.ComponentID == component {
			return item.Version
		}
	}
	return ""
}
