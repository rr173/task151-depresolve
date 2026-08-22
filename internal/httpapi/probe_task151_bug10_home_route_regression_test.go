package httpapi

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"depresolve/internal/service"
	"depresolve/internal/store"
)

func TestBug10_HomeRouteRedirectsToApplicationPage(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "resolver.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	svc, err := service.New(st)
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	New(svc).Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	if recorder.Code != http.StatusTemporaryRedirect || recorder.Header().Get("Location") != "/static/index.html" {
		t.Fatalf("status=%d location=%q", recorder.Code, recorder.Header().Get("Location"))
	}
}
