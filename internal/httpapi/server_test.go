package httpapi

import (
	"context"
	"depresolve/internal/model"
	"depresolve/internal/service"
	"depresolve/internal/store"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestStatsRoute(t *testing.T) {
	st, e := store.Open(filepath.Join(t.TempDir(), "x.db"))
	if e != nil {
		t.Fatal(e)
	}
	defer st.Close()
	svc, e := service.New(st)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = svc.CreateComponent(context.Background(), model.CreateComponentRequest{ID: "a", Name: "A", Ecosystem: "go"}); e != nil {
		t.Fatal(e)
	}
	rr := httptest.NewRecorder()
	New(svc).Handler().ServeHTTP(rr, httptest.NewRequest("GET", "/api/stats", nil))
	if rr.Code != 200 {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}

// Registering the same component twice is a business conflict and must surface
// to callers as 409 Conflict, not as an opaque 500 internal error.
func TestDuplicateComponentIsConflict(t *testing.T) {
	st, e := store.Open(filepath.Join(t.TempDir(), "x.db"))
	if e != nil {
		t.Fatal(e)
	}
	defer st.Close()
	svc, e := service.New(st)
	if e != nil {
		t.Fatal(e)
	}
	h := New(svc).Handler()
	body := `{"id":"app","name":"App","ecosystem":"go"}`
	first := httptest.NewRecorder()
	h.ServeHTTP(first, httptest.NewRequest("POST", "/api/components", strings.NewReader(body)))
	if first.Code != http.StatusCreated {
		t.Fatalf("first create status=%d body=%s", first.Code, first.Body.String())
	}
	second := httptest.NewRecorder()
	h.ServeHTTP(second, httptest.NewRequest("POST", "/api/components", strings.NewReader(body)))
	if second.Code != http.StatusConflict {
		t.Fatalf("duplicate create status=%d want=%d body=%s", second.Code, http.StatusConflict, second.Body.String())
	}
	if !strings.Contains(second.Body.String(), "app") {
		t.Fatalf("conflict body should identify the conflicting component: %s", second.Body.String())
	}
}

