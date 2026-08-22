package service

import (
	"context"
	"path/filepath"
	"testing"

	"depresolve/internal/model"
	"depresolve/internal/store"
)

func newTestService(t *testing.T) *Service {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "x.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	svc, err := New(st)
	if err != nil {
		t.Fatal(err)
	}
	return svc
}

func TestRetryReusesOriginalLocks(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)
	for _, c := range []model.CreateComponentRequest{{ID: "app", Name: "Application", Ecosystem: "go"}, {ID: "lib", Name: "Library", Ecosystem: "go"}} {
		if _, err := svc.CreateComponent(ctx, c); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := svc.CreateRelease(ctx, "app", model.CreateReleaseRequest{ID: "app-1", Version: "1.0.0", Platforms: []string{"linux/amd64"}, Dependencies: []model.DependencyInput{{ID: "app-lib", TargetComponent: "lib", Constraint: "^1.0.0"}}}); err != nil {
		t.Fatal(err)
	}
	for _, r := range []struct{ c, id, v string }{{"lib", "lib-1", "1.0.0"}, {"lib", "lib-2", "1.2.0"}} {
		if _, err := svc.CreateRelease(ctx, r.c, model.CreateReleaseRequest{ID: r.id, Version: r.v, Platforms: []string{"linux/amd64"}}); err != nil {
			t.Fatal(err)
		}
	}

	// Resolve pinned to lib 1.0.0 even though 1.2.0 is compatible.
	orig, err := svc.Resolve(ctx, model.ResolveRequest{ID: "orig", RootComponent: "app", RootConstraint: "*", Platform: "linux/amd64", Locks: map[string]string{"lib": "1.0.0"}, Policy: "locked"})
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	libOrig := selectionFor(orig.Selections, "lib")
	if libOrig.Version != "1.0.0" {
		t.Fatalf("orig selected lib=%s, want 1.0.0", libOrig.Version)
	}

	// Retry must reuse the original lock and policy, not drift to 1.2.0.
	retry, err := svc.RetryResolve(ctx, orig.Run.ID)
	if err != nil {
		t.Fatalf("retry: %v", err)
	}
	libRetry := selectionFor(retry.Selections, "lib")
	if libRetry.Version != "1.0.0" {
		t.Fatalf("retry drifted: selected lib=%s, want 1.0.0 (original lock ignored)", libRetry.Version)
	}
	if retry.Run.InputDigest != orig.Run.InputDigest {
		t.Fatalf("input digest changed: retry=%s orig=%s (locks/policy not carried over)", retry.Run.InputDigest, orig.Run.InputDigest)
	}
	for _, sel := range retry.Selections {
		if sel.ComponentID == "lib" && sel.Reason != "pinned by lock" {
			t.Fatalf("retry lib reason=%q, want %q", sel.Reason, "pinned by lock")
		}
	}
}

func selectionFor(sels []model.Selection, component string) model.Selection {
	for _, s := range sels {
		if s.ComponentID == component {
			return s
		}
	}
	return model.Selection{}
}
