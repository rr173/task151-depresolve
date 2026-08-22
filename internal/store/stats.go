package store

import (
	"context"

	"depresolve/internal/model"
)

func (s *Store) Stats(ctx context.Context) (model.Stats, error) {
	var x model.Stats
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM components`).Scan(&x.Components)
	if err != nil {
		return x, err
	}
	if err = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM releases`).Scan(&x.Releases); err != nil {
		return x, err
	}
	if err = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM resolve_runs WHERE status='resolved'`).Scan(&x.Resolved); err != nil {
		return x, err
	}
	if err = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM conflicts`).Scan(&x.Conflicts); err != nil {
		return x, err
	}
	if err = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM snapshots WHERE status='active'`).Scan(&x.ActiveSnapshots); err != nil {
		return x, err
	}
	return x, nil
}
