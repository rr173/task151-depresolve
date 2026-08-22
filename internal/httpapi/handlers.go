package httpapi

import (
	"depresolve/internal/model"
	"net/http"
	"strconv"
	"time"
)

func (s *Server) home(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/static/index.html", http.StatusTemporaryRedirect)
}
func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "uptime_seconds": int64(timeSince(s.started))})
}
func timeSince(t time.Time) float64 { return time.Since(t).Seconds() }
func (s *Server) ready(w http.ResponseWriter, r *http.Request) {
	if _, err := s.svc.Stats(r.Context()); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ready": true})
}
func (s *Server) stats(w http.ResponseWriter, r *http.Request) {
	x, err := s.svc.Stats(r.Context())
	if err != nil {
		writeErr(w, err)
		return
	}
	m := s.Metrics()
	writeJSON(w, http.StatusOK, map[string]any{"components": x.Components, "releases": x.Releases, "resolved": x.Resolved, "conflicts": x.Conflicts, "active_snapshots": x.ActiveSnapshots, "metrics": m})
}
func (s *Server) createComponent(w http.ResponseWriter, r *http.Request) {
	var req model.CreateComponentRequest
	if err := decode(r, &req); err != nil {
		writeErr(w, model.ErrInvalidInput)
		return
	}
	x, err := s.svc.CreateComponent(r.Context(), req)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, x)
}
func (s *Server) listComponents(w http.ResponseWriter, r *http.Request) {
	x, err := s.svc.ListComponents(r.Context())
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": x})
}
func (s *Server) getComponent(w http.ResponseWriter, r *http.Request) {
	x, err := s.svc.GetComponent(r.Context(), r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, x)
}
func (s *Server) deleteComponent(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteComponent(r.Context(), r.PathValue("id")); err != nil {
		writeErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func (s *Server) createRelease(w http.ResponseWriter, r *http.Request) {
	var req model.CreateReleaseRequest
	if err := decode(r, &req); err != nil {
		writeErr(w, model.ErrInvalidInput)
		return
	}
	x, err := s.svc.CreateRelease(r.Context(), r.PathValue("id"), req)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, x)
}
func (s *Server) listReleases(w http.ResponseWriter, r *http.Request) {
	x, err := s.svc.ListReleases(r.Context(), r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": x})
}
func (s *Server) getRelease(w http.ResponseWriter, r *http.Request) {
	x, err := s.svc.GetRelease(r.Context(), r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, x)
}
func (s *Server) resolve(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.listResolves(w, r)
		return
	}
	var req model.ResolveRequest
	if err := decode(r, &req); err != nil {
		writeErr(w, model.ErrInvalidInput)
		return
	}
	x, err := s.svc.Resolve(r.Context(), req)
	if err != nil {
		writeJSON(w, http.StatusConflict, x)
		return
	}
	writeJSON(w, http.StatusCreated, x)
}
func (s *Server) listResolves(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	x, err := s.svc.ListResolves(r.Context(), limit, offset)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, x)
}
func (s *Server) getResolve(w http.ResponseWriter, r *http.Request) {
	x, err := s.svc.GetResolve(r.Context(), r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, x)
}
func (s *Server) getGraph(w http.ResponseWriter, r *http.Request) {
	x, err := s.svc.GetResolve(r.Context(), r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"selections": x.Selections, "edges": x.Edges})
}
func (s *Server) getConflicts(w http.ResponseWriter, r *http.Request) {
	x, err := s.svc.GetResolve(r.Context(), r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"conflicts": x.Conflicts})
}
func (s *Server) retry(w http.ResponseWriter, r *http.Request) {
	x, err := s.svc.RetryResolve(r.Context(), r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, x)
}
func (s *Server) createSnapshot(w http.ResponseWriter, r *http.Request) {
	var req model.CreateSnapshotRequest
	_ = decode(r, &req)
	x, err := s.svc.CreateSnapshot(r.Context(), r.PathValue("id"), req)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, x)
}
func (s *Server) listSnapshots(w http.ResponseWriter, r *http.Request) {
	x, err := s.svc.ListSnapshots(r.Context())
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": x})
}
func (s *Server) getSnapshot(w http.ResponseWriter, r *http.Request) {
	x, items, err := s.svc.GetSnapshot(r.Context(), r.PathValue("id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"snapshot": x, "items": items})
}
func (s *Server) activate(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.ActivateSnapshot(r.Context(), r.PathValue("id")); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"activated": true})
}
func (s *Server) rollback(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.RollbackSnapshot(r.Context(), r.PathValue("id")); err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"rolled_back": true})
}
func (s *Server) diff(w http.ResponseWriter, r *http.Request) {
	x, err := s.svc.DiffSnapshots(r.Context(), r.URL.Query().Get("from"), r.URL.Query().Get("to"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, x)
}
func (s *Server) audit(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	x, err := s.svc.Audit(r.Context(), limit)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": x})
}
func (s *Server) rebuild(w http.ResponseWriter, r *http.Request) {
	x, err := s.svc.Recover(r.Context())
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, x)
}
func (s *Server) capabilities(w http.ResponseWriter, r *http.Request) {
	cat := s.svc.CatalogSnapshot()
	writeJSON(w, http.StatusOK, map[string]any{"capabilities": cat.Capabilities})
}
