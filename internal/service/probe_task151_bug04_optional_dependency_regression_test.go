package service

import (
	"context"
	"path/filepath"
	"testing"

	"depresolve/internal/model"
	"depresolve/internal/store"
)

func TestBug04_OptionalMissingDependencyDoesNotBlockResolve(t *testing.T) {
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
	if _, err = svc.CreateComponent(ctx, model.CreateComponentRequest{ID: "app", Name: "App", Ecosystem: "go"}); err != nil {
		t.Fatal(err)
	}
	if _, err = svc.CreateComponent(ctx, model.CreateComponentRequest{ID: "plugin", Name: "Plugin", Ecosystem: "go"}); err != nil {
		t.Fatal(err)
	}
	if _, err = svc.CreateRelease(ctx, "app", model.CreateReleaseRequest{ID: "app-1", Version: "1.0.0", Dependencies: []model.DependencyInput{{ID: "optional-plugin", TargetComponent: "plugin", Constraint: "^1.0.0", Optional: true}}}); err != nil {
		t.Fatal(err)
	}
	result, err := svc.Resolve(ctx, model.ResolveRequest{RootComponent: "app", RootConstraint: "*"})
	if err != nil {
		t.Fatalf("optional dependency blocked resolve: %v", err)
	}
	if len(result.Selections) != 1 || result.Selections[0].ComponentID != "app" {
		t.Fatalf("selections=%+v", result.Selections)
	}
}
