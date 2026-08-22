package service

import (
	"context"
	"fmt"

	"depresolve/internal/model"
	"depresolve/internal/solver"
)

func (s *Service) Resolve(ctx context.Context, req model.ResolveRequest) (model.ResolveResponse, error) {
	if req.ID == "" {
		req.ID = newID("run")
	}
	if req.RootConstraint == "" {
		req.RootConstraint = "*"
	}
	if req.Platform == "" {
		req.Platform = "any"
	}
	lock := s.keyLock(req.RootComponent)
	lock.Lock()
	defer lock.Unlock()
	now := clock()
	run := model.ResolveRun{ID: req.ID, RootComponent: req.RootComponent, RootConstraint: req.RootConstraint, Platform: req.Platform, AllowPrerelease: req.AllowPrerelease, InputDigest: solver.InputDigest(req), Status: model.RunPending, CreatedAt: now}
	if err := s.store.CreateRun(ctx, run, req); err != nil {
		return model.ResolveResponse{}, err
	}
	s.counters.IncRuns()
	s.mu.RLock()
	cat := s.catalog
	s.mu.RUnlock()
	result, err := solver.Solve(solver.Input{Request: req, Catalog: cat})
	if err != nil {
		run.Status = model.RunFailed
		run.Error = err.Error()
		run.FinishedAt = &now
		_ = s.store.SaveResult(ctx, run.ID, result.Selections, result.Edges, result.Conflicts)
		_ = s.store.UpdateRun(ctx, run)
		s.counters.IncFailures()
		return model.ResolveResponse{Run: run, Selections: result.Selections, Edges: result.Edges, Conflicts: result.Conflicts}, err
	}
	run.Status = model.RunResolved
	run.FinishedAt = &now
	if err := s.store.SaveResult(ctx, run.ID, result.Selections, result.Edges, result.Conflicts); err != nil {
		return model.ResolveResponse{}, err
	}
	if err := s.store.UpdateRun(ctx, run); err != nil {
		return model.ResolveResponse{}, err
	}
	s.counters.IncResolved()
	_ = s.store.Audit(ctx, "resolve", run.ID, "resolve", fmt.Sprintf("selected=%d", len(result.Selections)))
	return model.ResolveResponse{Run: run, Selections: result.Selections, Edges: result.Edges, Conflicts: result.Conflicts}, nil
}
func (s *Service) GetResolve(ctx context.Context, id string) (model.ResolveResponse, error) {
	run, err := s.store.GetRun(ctx, id)
	if err != nil {
		return model.ResolveResponse{}, err
	}
	sel, err := s.store.GetSelections(ctx, id)
	if err != nil {
		return model.ResolveResponse{}, err
	}
	edges, err := s.store.GetEdges(ctx, id)
	if err != nil {
		return model.ResolveResponse{}, err
	}
	conflicts, err := s.store.GetConflicts(ctx, id)
	return model.ResolveResponse{Run: run, Selections: sel, Edges: edges, Conflicts: conflicts}, err
}
func (s *Service) ListResolves(ctx context.Context, limit, offset int) (model.Page[model.ResolveRun], error) {
	items, total, err := s.store.ListRuns(ctx, limit, offset)
	return model.Page[model.ResolveRun]{Items: items, Limit: limit, Offset: offset, Total: total}, err
}
func (s *Service) RetryResolve(ctx context.Context, id string) (model.ResolveResponse, error) {
	r, err := s.store.GetRun(ctx, id)
	if err != nil {
		return model.ResolveResponse{}, err
	}
	return s.Resolve(ctx, model.ResolveRequest{ID: newID("retry"), RootComponent: r.RootComponent, RootConstraint: r.RootConstraint, Platform: r.Platform, AllowPrerelease: r.AllowPrerelease})
}
