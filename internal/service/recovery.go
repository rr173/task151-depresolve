package service

import (
	"context"
	"depresolve/internal/model"
)

type RecoveryReport struct {
	Components int    `json:"components"`
	Releases   int    `json:"releases"`
	Active     string `json:"active_snapshot,omitempty"`
	Healthy    bool   `json:"healthy"`
}

func (s *Service) Recover(ctx context.Context) (RecoveryReport, error) {
	if err := s.RebuildIndex(ctx); err != nil {
		return RecoveryReport{}, err
	}
	st, err := s.store.Stats(ctx)
	if err != nil {
		return RecoveryReport{}, err
	}
	out := RecoveryReport{Components: st.Components, Releases: st.Releases, Healthy: true}
	if active, e := s.store.ActiveSnapshot(ctx); e == nil {
		out.Active = active.ID
	}
	return out, nil
}
func (s *Service) Stats(ctx context.Context) (model.Stats, error) { return s.store.Stats(ctx) }
func (s *Service) Audit(ctx context.Context, limit int) ([]model.AuditEvent, error) {
	return s.store.ListAudit(ctx, limit)
}
