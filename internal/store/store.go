package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"depresolve/internal/model"
	_ "modernc.org/sqlite"
)

type Store struct{ db *sql.DB }

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	s := &Store{db: db}
	if _, err = db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("schema: %w", err)
	}
	return s, nil
}
func (s *Store) Close() error { return s.db.Close() }
func (s *Store) DB() *sql.DB  { return s.db }
func nowUnix() int64          { return time.Now().UTC().Unix() }
func marshal(v any) string    { b, _ := json.Marshal(v); return string(b) }
func unmarshal(raw string, v any) error {
	if raw == "" {
		raw = "[]"
	}
	return json.Unmarshal([]byte(raw), v)
}
func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
func intBool(v int) bool { return v != 0 }

func (s *Store) CreateComponent(ctx context.Context, c model.Component) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO components(id,name,ecosystem,stable_only,created_at,updated_at) VALUES(?,?,?,?,?,?)`, c.ID, c.Name, c.Ecosystem, boolInt(c.StableOnly), c.CreatedAt.Unix(), c.UpdatedAt.Unix())
	if err != nil && strings.Contains(err.Error(), "UNIQUE") {
		return model.ErrAlreadyExists
	}
	return err
}
func (s *Store) ListComponents(ctx context.Context) ([]model.Component, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,name,ecosystem,stable_only,created_at,updated_at FROM components ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Component{}
	for rows.Next() {
		var c model.Component
		var stable int64
		var created, updated int64
		if err := rows.Scan(&c.ID, &c.Name, &c.Ecosystem, &stable, &created, &updated); err != nil {
			return nil, err
		}
		c.StableOnly = intBool(int(stable))
		c.CreatedAt = time.Unix(created, 0).UTC()
		c.UpdatedAt = time.Unix(updated, 0).UTC()
		out = append(out, c)
	}
	return out, rows.Err()
}
func (s *Store) GetComponent(ctx context.Context, id string) (model.Component, error) {
	var c model.Component
	var stable int64
	var created, updated int64
	err := s.db.QueryRowContext(ctx, `SELECT id,name,ecosystem,stable_only,created_at,updated_at FROM components WHERE id=?`, id).Scan(&c.ID, &c.Name, &c.Ecosystem, &stable, &created, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return c, model.ErrNotFound
	}
	if err != nil {
		return c, err
	}
	c.StableOnly = intBool(int(stable))
	c.CreatedAt = time.Unix(created, 0).UTC()
	c.UpdatedAt = time.Unix(updated, 0).UTC()
	return c, nil
}
func (s *Store) DeleteComponent(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM components WHERE id=?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return model.ErrNotFound
	}
	return nil
}

func (s *Store) CreateRelease(ctx context.Context, r model.Release) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO releases(id,component_id,version,platforms,capabilities,conflicts,published,created_at) VALUES(?,?,?,?,?,?,?,?)`, r.ID, r.ComponentID, r.Version, marshal(r.Platforms), marshal(r.Capabilities), marshal(r.Conflicts), boolInt(r.Published), r.CreatedAt.Unix()); err != nil {
		tx.Rollback()
		if strings.Contains(err.Error(), "UNIQUE") {
			return model.ErrAlreadyExists
		}
		return err
	}
	for _, d := range r.Dependencies {
		if d.ID == "" {
			d.ID = r.ID + "/" + d.TargetComponent
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO dependencies(id,source_release_id,target_component,constraint_text,optional,platform,locked) VALUES(?,?,?,?,?,?,?)`, d.ID, r.ID, d.TargetComponent, d.Constraint, boolInt(d.Optional), d.Platform, boolInt(d.Locked)); err != nil {
			tx.Rollback()
			return err
		}
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}
func (s *Store) ListReleases(ctx context.Context, component string) ([]model.Release, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,component_id,version,platforms,capabilities,conflicts,published,created_at FROM releases WHERE component_id=? ORDER BY version DESC,id`, component)
	if err != nil {
		return nil, err
	}
	out := []model.Release{}
	for rows.Next() {
		r, err := scanRelease(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()
	for i := range out {
		deps, err := s.listDependencies(ctx, out[i].ID)
		if err != nil {
			return nil, err
		}
		out[i].Dependencies = deps
	}
	return out, nil
}
func (s *Store) ListAllReleases(ctx context.Context) ([]model.Release, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,component_id,version,platforms,capabilities,conflicts,published,created_at FROM releases ORDER BY component_id,version DESC,id`)
	if err != nil {
		return nil, err
	}
	out := []model.Release{}
	for rows.Next() {
		r, err := scanRelease(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()
	for i := range out {
		deps, err := s.listDependencies(ctx, out[i].ID)
		if err != nil {
			return nil, err
		}
		out[i].Dependencies = deps
	}
	return out, nil
}
func (s *Store) GetRelease(ctx context.Context, id string) (model.Release, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id,component_id,version,platforms,capabilities,conflicts,published,created_at FROM releases WHERE id=?`, id)
	r, err := scanRelease(row)
	if errors.Is(err, sql.ErrNoRows) {
		return r, model.ErrNotFound
	}
	if err != nil {
		return r, err
	}
	r.Dependencies, err = s.listDependencies(ctx, id)
	return r, err
}

type scanner interface{ Scan(...any) error }

func scanRelease(x scanner) (model.Release, error) {
	var r model.Release
	var platforms, caps, conflicts string
	var published int
	var created int64
	if err := x.Scan(&r.ID, &r.ComponentID, &r.Version, &platforms, &caps, &conflicts, &published, &created); err != nil {
		return r, err
	}
	if err := unmarshal(platforms, &r.Platforms); err != nil {
		return r, err
	}
	if err := unmarshal(caps, &r.Capabilities); err != nil {
		return r, err
	}
	if err := unmarshal(conflicts, &r.Conflicts); err != nil {
		return r, err
	}
	r.Published = intBool(published)
	r.CreatedAt = time.Unix(created, 0).UTC()
	return r, nil
}
func (s *Store) listDependencies(ctx context.Context, releaseID string) ([]model.Dependency, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,source_release_id,target_component,constraint_text,optional,platform,locked FROM dependencies WHERE source_release_id=? ORDER BY id`, releaseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Dependency{}
	for rows.Next() {
		var d model.Dependency
		var optional, locked int
		if err := rows.Scan(&d.ID, &d.SourceReleaseID, &d.TargetComponent, &d.Constraint, &optional, &d.Platform, &locked); err != nil {
			return nil, err
		}
		d.Optional = !intBool(optional)
		d.Locked = intBool(locked)
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s *Store) CreateRun(ctx context.Context, r model.ResolveRun, req model.ResolveRequest) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO resolve_runs(id,root_component,root_constraint,platform,allow_prerelease,locks,policy,input_digest,status,error,created_at) VALUES(?,?,?,?,?,?,?,?,?,?,?)`, r.ID, r.RootComponent, r.RootConstraint, r.Platform, boolInt(r.AllowPrerelease), marshal(req.Locks), req.Policy, r.InputDigest, r.Status, r.Error, r.CreatedAt.Unix())
	return err
}
func (s *Store) UpdateRun(ctx context.Context, r model.ResolveRun) error {
	finished := int64(0)
	if r.FinishedAt != nil {
		finished = r.FinishedAt.Unix()
	}
	_, err := s.db.ExecContext(ctx, `UPDATE resolve_runs SET status=?,error=?,finished_at=? WHERE id=?`, r.Status, r.Error, finished, r.ID)
	return err
}
func (s *Store) GetRun(ctx context.Context, id string) (model.ResolveRun, error) {
	var r model.ResolveRun
	var allow int
	var created, finished int64
	var locks string
	err := s.db.QueryRowContext(ctx, `SELECT id,root_component,root_constraint,platform,allow_prerelease,locks,policy,input_digest,status,error,created_at,finished_at FROM resolve_runs WHERE id=?`, id).Scan(&r.ID, &r.RootComponent, &r.RootConstraint, &r.Platform, &allow, &locks, &r.Policy, &r.InputDigest, &r.Status, &r.Error, &created, &finished)
	if errors.Is(err, sql.ErrNoRows) {
		return r, model.ErrNotFound
	}
	if err != nil {
		return r, err
	}
	r.AllowPrerelease = intBool(allow)
	if err = unmarshal(locks, &r.Locks); err != nil {
		return r, err
	}
	r.CreatedAt = time.Unix(created, 0).UTC()
	if finished != 0 {
		x := time.Unix(finished, 0).UTC()
		r.FinishedAt = &x
	}
	return r, nil
}
func (s *Store) ListRuns(ctx context.Context, limit, offset int) ([]model.ResolveRun, int, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id,root_component,root_constraint,platform,allow_prerelease,locks,policy,input_digest,status,error,created_at,finished_at FROM resolve_runs ORDER BY created_at DESC,id DESC LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []model.ResolveRun{}
	for rows.Next() {
		var r model.ResolveRun
		var allow int
		var created, finished int64
		var locks string
		if err := rows.Scan(&r.ID, &r.RootComponent, &r.RootConstraint, &r.Platform, &allow, &locks, &r.Policy, &r.InputDigest, &r.Status, &r.Error, &created, &finished); err != nil {
			return nil, 0, err
		}
		r.AllowPrerelease = intBool(allow)
		if err := unmarshal(locks, &r.Locks); err != nil {
			return nil, 0, err
		}
		r.CreatedAt = time.Unix(created, 0).UTC()
		if finished != 0 {
			x := time.Unix(finished, 0).UTC()
			r.FinishedAt = &x
		}
		out = append(out, r)
	}
	var total int
	err = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM resolve_runs`).Scan(&total)
	return out, total, err
}

func (s *Store) SaveResult(ctx context.Context, runID string, selections []model.Selection, edges []model.Edge, conflicts []model.Conflict) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	for _, x := range selections {
		if _, err = tx.ExecContext(ctx, `INSERT OR REPLACE INTO selections(run_id,component_id,release_id,version,reason,score) VALUES(?,?,?,?,?,?)`, runID, x.ComponentID, x.ReleaseID, x.Version, x.Reason, x.Score); err != nil {
			tx.Rollback()
			return err
		}
	}
	for _, x := range edges {
		if _, err = tx.ExecContext(ctx, `INSERT INTO edges(run_id,from_component,to_component,constraint_text,optional,satisfied,path) VALUES(?,?,?,?,?,?,?)`, runID, x.FromComponent, x.ToComponent, x.Constraint, boolInt(x.Optional), boolInt(x.Satisfied), x.Path); err != nil {
			tx.Rollback()
			return err
		}
	}
	for i, x := range conflicts {
		if x.ID == "" {
			x.ID = fmt.Sprintf("%s-c%d", runID, i)
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO conflicts(id,run_id,component_id,kind,message,path,candidates) VALUES(?,?,?,?,?,?,?)`, x.ID, runID, x.ComponentID, x.Kind, x.Message, x.Path, marshal(x.Candidates)); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}
func (s *Store) GetSelections(ctx context.Context, runID string) ([]model.Selection, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT component_id,release_id,version,reason,score FROM selections WHERE run_id=? ORDER BY component_id`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Selection{}
	for rows.Next() {
		var x model.Selection
		x.RunID = runID
		if err := rows.Scan(&x.ComponentID, &x.ReleaseID, &x.Version, &x.Reason, &x.Score); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func (s *Store) GetEdges(ctx context.Context, runID string) ([]model.Edge, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT from_component,to_component,constraint_text,optional,satisfied,path FROM edges WHERE run_id=? ORDER BY from_component,to_component`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Edge{}
	for rows.Next() {
		var x model.Edge
		var optional, satisfied int
		x.RunID = runID
		if err := rows.Scan(&x.FromComponent, &x.ToComponent, &x.Constraint, &optional, &satisfied, &x.Path); err != nil {
			return nil, err
		}
		x.Optional = intBool(optional)
		x.Satisfied = intBool(satisfied)
		out = append(out, x)
	}
	return out, rows.Err()
}
func (s *Store) GetConflicts(ctx context.Context, runID string) ([]model.Conflict, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,component_id,kind,message,path,candidates FROM conflicts WHERE run_id=? ORDER BY id`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Conflict{}
	for rows.Next() {
		var x model.Conflict
		var candidates string
		x.RunID = runID
		if err := rows.Scan(&x.ID, &x.ComponentID, &x.Kind, &x.Message, &x.Path, &candidates); err != nil {
			return nil, err
		}
		if err := unmarshal(candidates, &x.Candidates); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
