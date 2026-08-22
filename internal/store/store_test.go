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
