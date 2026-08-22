package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"depresolve/internal/model"
)

func (s *Store) CreateSnapshot(ctx context.Context, snap model.Snapshot, items []model.SnapshotItem) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO snapshots(id,run_id,name,status,digest,created_at) VALUES(?,?,?,?,?,?)`, snap.ID, snap.RunID, snap.Name, snap.Status, snap.Digest, snap.CreatedAt.Unix()); err != nil {
		tx.Rollback()
		return err
	}
	for _, x := range items {
		if _, err = tx.ExecContext(ctx, `INSERT INTO snapshot_items(snapshot_id,component_id,release_id,version) VALUES(?,?,?,?)`, snap.ID, x.ComponentID, x.ReleaseID, x.Version); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}
func scanSnapshot(row interface{ Scan(...any) error }) (model.Snapshot, error) {
	var x model.Snapshot
	var created, activated int64
	if err := row.Scan(&x.ID, &x.RunID, &x.Name, &x.Status, &x.Digest, &created, &activated); err != nil {
		return x, err
	}
	x.CreatedAt = time.Unix(created, 0).UTC()
	if activated != 0 {
		t := time.Unix(activated, 0).UTC()
		x.ActivatedAt = &t
	}
	return x, nil
}
func (s *Store) GetSnapshot(ctx context.Context, id string) (model.Snapshot, error) {
	x, err := scanSnapshot(s.db.QueryRowContext(ctx, `SELECT id,run_id,name,status,digest,created_at,activated_at FROM snapshots WHERE id=?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return x, model.ErrNotFound
	}
	return x, err
}
func (s *Store) ListSnapshots(ctx context.Context) ([]model.Snapshot, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,run_id,name,status,digest,created_at,activated_at FROM snapshots ORDER BY created_at DESC,id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Snapshot{}
	for rows.Next() {
		x, err := scanSnapshot(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func (s *Store) GetSnapshotItems(ctx context.Context, id string) ([]model.SnapshotItem, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT snapshot_id,component_id,release_id,version FROM snapshot_items WHERE snapshot_id=? ORDER BY component_id`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.SnapshotItem{}
	for rows.Next() {
		var x model.SnapshotItem
		if err := rows.Scan(&x.SnapshotID, &x.ComponentID, &x.ReleaseID, &x.Version); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func (s *Store) ActiveSnapshot(ctx context.Context) (model.Snapshot, error) {
	x, err := scanSnapshot(s.db.QueryRowContext(ctx, `SELECT id,run_id,name,status,digest,created_at,activated_at FROM snapshots WHERE status='active'`))
	if errors.Is(err, sql.ErrNoRows) {
		return x, model.ErrNotFound
	}
	return x, err
}
func (s *Store) ActivateSnapshot(ctx context.Context, id string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	now := nowUnix()
	if _, err = tx.ExecContext(ctx, `UPDATE snapshots SET status='retired' WHERE status='active'`); err != nil {
		tx.Rollback()
		return err
	}
	res, err := tx.ExecContext(ctx, `UPDATE snapshots SET status='active',activated_at=? WHERE id=? AND status IN ('draft','retired')`, now, id)
	if err != nil {
		tx.Rollback()
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		tx.Rollback()
		return fmt.Errorf("%w: snapshot cannot activate", model.ErrState)
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO audit_events(aggregate,aggregate_id,action,summary,created_at) VALUES(?,?,?,?,?)`, "snapshot", id, "activate", "snapshot activated", now); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}
func (s *Store) RollbackSnapshot(ctx context.Context, id string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE snapshots SET status='rolled_back' WHERE id=? AND status='active'`, id); err != nil {
		tx.Rollback()
		return err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO audit_events(aggregate,aggregate_id,action,summary,created_at) VALUES(?,?,?,?,?)`, "snapshot", id, "rollback", "snapshot rolled back", nowUnix()); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}
