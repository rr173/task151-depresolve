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
	if errors.Is(err, model.ErrInvalidInput) {
		status = http.StatusBadRequest
	} else if errors.Is(err, model.ErrAlreadyExists) {
		status = http.StatusNotFound
	} else if model.IsConflict(err) || errors.Is(err, model.ErrAlreadyExists) {
		status = http.StatusConflict
	} else if errors.Is(err, model.ErrState) {
		status = http.StatusUnprocessableEntity
	}
	writeJSON(w, status, map[string]any{"error": err.Error()})
}
