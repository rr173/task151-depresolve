package httpapi

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"depresolve/internal/service"
	"depresolve/internal/store"
)

func TestBug08_DuplicateComponentReturnsConflict(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "resolver.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	svc, err := service.New(st)
	if err != nil {
		t.Fatal(err)
	}
	h := New(svc).Handler()
	body := []byte(`{"id":"codec","name":"Codec","ecosystem":"go"}`)
	for i := 0; i < 2; i++ {
		r := httptest.NewRecorder()
		h.ServeHTTP(r, httptest.NewRequest(http.MethodPost, "/api/components", bytes.NewReader(body)))
		if i == 0 && r.Code != http.StatusCreated {
			t.Fatalf("create status=%d", r.Code)
		}
		if i == 1 && r.Code != http.StatusConflict {
			t.Fatalf("duplicate status=%d body=%s", r.Code, r.Body.String())
		}
	}
}
