package service

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"depresolve/internal/model"
	"depresolve/internal/store"
)

func TestBug07_ConflictingCapabilitiesRejectResolve(t *testing.T) {
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
	for _, component := range []model.CreateComponentRequest{{ID: "app", Name: "App", Ecosystem: "go"}, {ID: "crypto", Name: "Crypto", Ecosystem: "go"}} {
		if _, err := svc.CreateComponent(ctx, component); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = svc.CreateRelease(ctx, "app", model.CreateReleaseRequest{ID: "app-1", Version: "1.0.0", Capabilities: []string{"fips"}, Dependencies: []model.DependencyInput{{ID: "app-crypto", TargetComponent: "crypto", Constraint: "^1.0.0"}}}); err != nil {
		t.Fatal(err)
	}
	if _, err = svc.CreateRelease(ctx, "crypto", model.CreateReleaseRequest{ID: "crypto-1", Version: "1.0.0", Conflicts: []string{"fips"}}); err != nil {
		t.Fatal(err)
	}
	_, err = svc.Resolve(ctx, model.ResolveRequest{RootComponent: "app", RootConstraint: "*"})
	if !errors.Is(err, model.ErrConflict) {
		t.Fatalf("capability conflict was accepted: %v", err)
	}
}
