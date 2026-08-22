package store

import (
	"context"
	"depresolve/internal/model"
	"time"
)

func (s *Store) Audit(ctx context.Context, aggregate, id, action, summary string) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO audit_events(aggregate,aggregate_id,action,summary,created_at) VALUES(?,?,?,?,?)`, aggregate, id, action, summary, nowUnix())
	return err
}
func (s *Store) ListAudit(ctx context.Context, limit int) ([]model.AuditEvent, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id,aggregate,aggregate_id,action,summary,created_at FROM audit_events ORDER BY id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.AuditEvent{}
	for rows.Next() {
		var x model.AuditEvent
		var ts int64
		if err := rows.Scan(&x.ID, &x.Aggregate, &x.AggregateID, &x.Action, &x.Summary, &ts); err != nil {
			return nil, err
		}
		x.CreatedAt = time.Unix(ts, 0).UTC()
		out = append(out, x)
	}
	return out, rows.Err()
}
