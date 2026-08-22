package httpapi

import (
	"net/http"
	"time"

	"depresolve/internal/metrics"
	"depresolve/internal/service"
	"depresolve/internal/webui"
)

type Server struct {
	svc     *service.Service
	mux     *http.ServeMux
	started time.Time
}

func New(svc *service.Service) *Server {
	s := &Server{svc: svc, mux: http.NewServeMux(), started: time.Now().UTC()}
	s.routes()
	return s
}
func (s *Server) Handler() http.Handler { return s.mux }
func (s *Server) routes() {
	s.mux.Handle("GET /", http.HandlerFunc(s.home))
	s.mux.Handle("GET /static/", http.StripPrefix("/static/", webui.Handler()))
	s.mux.HandleFunc("GET /healthz", s.health)
	s.mux.HandleFunc("GET /readyz", s.ready)
	s.mux.HandleFunc("GET /api/stats", s.stats)
	s.mux.HandleFunc("POST /api/components", s.createComponent)
	s.mux.HandleFunc("GET /api/components", s.listComponents)
	s.mux.HandleFunc("GET /api/components/{id}", s.getComponent)
	s.mux.HandleFunc("DELETE /api/components/{id}", s.deleteComponent)
	s.mux.HandleFunc("POST /api/components/{id}/releases", s.createRelease)
	s.mux.HandleFunc("GET /api/components/{id}/releases", s.listReleases)
	s.mux.HandleFunc("GET /api/releases/{id}", s.getRelease)
	s.mux.HandleFunc("POST /api/resolve", s.resolve)
	s.mux.HandleFunc("GET /api/resolve", s.listResolves)
	s.mux.HandleFunc("GET /api/resolve/{id}", s.getResolve)
	s.mux.HandleFunc("GET /api/resolve/{id}/graph", s.getGraph)
	s.mux.HandleFunc("GET /api/resolve/{id}/conflicts", s.getConflicts)
	s.mux.HandleFunc("POST /api/resolve/{id}/retry", s.retry)
	s.mux.HandleFunc("POST /api/resolve/{id}/snapshot", s.createSnapshot)
	s.mux.HandleFunc("GET /api/snapshots", s.listSnapshots)
	s.mux.HandleFunc("GET /api/snapshots/{id}", s.getSnapshot)
	s.mux.HandleFunc("POST /api/snapshots/{id}/activate", s.activate)
	s.mux.HandleFunc("POST /api/snapshots/{id}/rollback", s.rollback)
	s.mux.HandleFunc("GET /api/snapshots/diff", s.diff)
	s.mux.HandleFunc("GET /api/audit", s.audit)
	s.mux.HandleFunc("POST /api/rebuild", s.rebuild)
	s.mux.HandleFunc("GET /api/capabilities", s.capabilities)
}
func (s *Server) Metrics() metrics.Snapshot { return s.svc.Counters().Snapshot() }
