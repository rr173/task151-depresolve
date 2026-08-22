package httpapi

import (
	"context"
	"depresolve/internal/model"
	"depresolve/internal/service"
	"depresolve/internal/store"
	"net/http/httptest"
	"path/filepath"
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
