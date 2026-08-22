package service

import (
	"context"
	"fmt"
	"depresolve/internal/model"
	"depresolve/internal/plan"
)

func (s *Service) CreateSnapshot(ctx context.Context, runID string, req model.CreateSnapshotRequest) (model.Snapshot, error) {
	if req.Name == "" {
		req.Name = "snapshot-" + runID
	}
	out, err := s.GetResolve(ctx, runID)
	if err != nil {
		return model.Snapshot{}, err
	}
	if out.Run.Status != model.RunResolved {
		return model.Snapshot{}, fmt.Errorf("%w: resolve is not successful", model.ErrState)
	}
	items := make([]model.SnapshotItem, 0, len(out.Selections))
	for _, x := range out.Selections {
		items = append(items, model.SnapshotItem{SnapshotID: runID, ComponentID: x.ComponentID, ReleaseID: x.ReleaseID, Version: x.Version})
	}
	snap := model.Snapshot{ID: newID("snap"), RunID: runID, Name: req.Name, Status: model.SnapshotDraft, CreatedAt: clock()}
	snap.Digest = plan.Digest(items)
	for i := range items {
		items[i].SnapshotID = snap.ID
	}
	if err := s.store.CreateSnapshot(ctx, snap, items); err != nil {
		return snap, err
	}
	s.counters.IncSnapshots()
	return snap, nil
}
func (s *Service) ListSnapshots(ctx context.Context) ([]model.Snapshot, error) {
	return s.store.ListSnapshots(ctx)
}
func (s *Service) GetSnapshot(ctx context.Context, id string) (model.Snapshot, []model.SnapshotItem, error) {
	x, err := s.store.GetSnapshot(ctx, id)
	if err != nil {
		return x, nil, err
	}
	items, err := s.store.GetSnapshotItems(ctx, id)
	return x, items, err
}
func (s *Service) ActivateSnapshot(ctx context.Context, id string) error {
	if err := s.store.ActivateSnapshot(ctx, id); err != nil {
		return err
	}
	s.counters.IncActivations()
	return nil
}
func (s *Service) RollbackSnapshot(ctx context.Context, id string) error {
	return s.store.RollbackSnapshot(ctx, id)
}
func (s *Service) DiffSnapshots(ctx context.Context, from, to string) (model.SnapshotDiff, error) {
	a, err := s.store.GetSnapshotItems(ctx, from)
	if err != nil {
		return model.SnapshotDiff{}, err
	}
	b, err := s.store.GetSnapshotItems(ctx, to)
	if err != nil {
		return model.SnapshotDiff{}, err
	}
	return plan.Diff(plan.SelectionMap(a), plan.SelectionMap(b)), nil
}
