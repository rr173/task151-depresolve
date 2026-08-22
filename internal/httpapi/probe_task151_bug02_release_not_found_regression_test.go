package httpapi

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"depresolve/internal/service"
	"depresolve/internal/store"
)

func TestBug02_MissingReleaseReturnsNotFound(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "resolver.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	svc, err := service.New(st)
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	New(svc).Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/releases/missing", nil))
	if response.Code != http.StatusNotFound {
		t.Fatalf("missing release status=%d body=%s", response.Code, response.Body.String())
	}
}
