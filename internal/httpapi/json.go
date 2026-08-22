package httpapi

import (
	"encoding/json"
	"errors"
	"depresolve/internal/model"
	"io"
	"net/http"
)

func decode(r *http.Request, v any) error {
	if r.Body == nil {
		return model.ErrInvalidInput
	}
	d := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	d.DisallowUnknownFields()
	if err := d.Decode(v); err != nil {
		return err
	}
	return nil
}
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func writeErr(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, model.ErrInvalidInput):
		status = http.StatusBadRequest
	case model.IsNotFound(err):
		status = http.StatusNotFound
	case model.IsConflict(err):
		status = http.StatusConflict
	case errors.Is(err, model.ErrState):
		status = http.StatusUnprocessableEntity
	}
	writeJSON(w, status, map[string]any{"error": err.Error()})
}
