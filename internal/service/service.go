package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"depresolve/internal/catalog"
	"depresolve/internal/metrics"
	"depresolve/internal/model"
	"depresolve/internal/store"
)

type Service struct {
	store    *store.Store
	mu       sync.RWMutex
	catalog  *catalog.Catalog
	locks    sync.Map
	counters *metrics.Counters
}

func New(st *store.Store) (*Service, error) {
	s := &Service{store: st, counters: &metrics.Counters{}}
	if err := s.RebuildIndex(context.Background()); err != nil {
		return nil, err
	}
	return s, nil
}
func (s *Service) Store() *store.Store         { return s.store }
func (s *Service) Counters() *metrics.Counters { return s.counters }
func (s *Service) keyLock(key string) *sync.Mutex {
	v, _ := s.locks.LoadOrStore(key, &sync.Mutex{})
	return v.(*sync.Mutex)
}
func newID(prefix string) string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
	}
	return prefix + "-" + hex.EncodeToString(b)
}
func clock() time.Time { return time.Now().UTC() }

func (s *Service) RebuildIndex(ctx context.Context) error {
	components, err := s.store.ListComponents(ctx)
	if err != nil {
		return err
	}
	releases, err := s.store.ListAllReleases(ctx)
	if err != nil {
		return err
	}
	idx, err := catalog.New(components, releases)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.catalog = idx
	s.mu.Unlock()
	return nil
}
func (s *Service) CatalogSnapshot() *catalog.Catalog {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.catalog
}

func (s *Service) CreateComponent(ctx context.Context, req model.CreateComponentRequest) (model.Component, error) {
	if req.ID == "" || req.Name == "" {
		return model.Component{}, model.ErrInvalidInput
	}
	now := clock()
	c := req.Normalize()
	c.CreatedAt = now
	c.UpdatedAt = now
	if err := s.store.CreateComponent(ctx, c); err != nil {
		return c, err
	}
	if err := s.RebuildIndex(ctx); err != nil {
		return c, err
	}
	_ = s.store.Audit(ctx, "component", c.ID, "create", c.Name)
	return c, nil
}
func (s *Service) ListComponents(ctx context.Context) ([]model.Component, error) {
	return s.store.ListComponents(ctx)
}
func (s *Service) GetComponent(ctx context.Context, id string) (model.Component, error) {
	return s.store.GetComponent(ctx, id)
}
func (s *Service) DeleteComponent(ctx context.Context, id string) error {
	if err := s.store.DeleteComponent(ctx, id); err != nil {
		return err
	}
	return s.RebuildIndex(ctx)
}
func (s *Service) CreateRelease(ctx context.Context, component string, req model.CreateReleaseRequest) (model.Release, error) {
	if req.Version == "" {
		return model.Release{}, model.ErrInvalidInput
	}
	if _, err := s.store.GetComponent(ctx, component); err != nil {
		return model.Release{}, err
	}
	r := model.Release{ID: req.ID, ComponentID: component, Version: req.Version, Platforms: req.Platforms, Capabilities: req.Capabilities, Conflicts: req.Conflicts, Published: true, CreatedAt: clock()}
	if r.ID == "" {
		r.ID = newID("rel")
	}
	for _, d := range req.Dependencies {
		r.Dependencies = append(r.Dependencies, model.Dependency{ID: d.ID, SourceReleaseID: r.ID, TargetComponent: d.TargetComponent, Constraint: d.Constraint, Optional: d.Optional, Platform: d.Platform, Locked: d.Locked})
	}
	if err := s.store.CreateRelease(ctx, r); err != nil {
		return r, err
	}
	if err := s.RebuildIndex(ctx); err != nil {
		return r, err
	}
	_ = s.store.Audit(ctx, "release", r.ID, "create", r.ComponentID+"@"+r.Version)
	return r, nil
}
func (s *Service) ListReleases(ctx context.Context, component string) ([]model.Release, error) {
	return s.store.ListReleases(ctx, component)
}
func (s *Service) GetRelease(ctx context.Context, id string) (model.Release, error) {
	return s.store.GetRelease(ctx, id)
}
