package store

const schema = `
PRAGMA foreign_keys = ON;
CREATE TABLE IF NOT EXISTS components(id TEXT PRIMARY KEY, name TEXT NOT NULL, ecosystem TEXT NOT NULL, stable_only INTEGER NOT NULL, created_at INTEGER NOT NULL, updated_at INTEGER NOT NULL);
CREATE TABLE IF NOT EXISTS releases(id TEXT PRIMARY KEY, component_id TEXT NOT NULL REFERENCES components(id) ON DELETE CASCADE, version TEXT NOT NULL, platforms TEXT NOT NULL, capabilities TEXT NOT NULL, conflicts TEXT NOT NULL, published INTEGER NOT NULL, created_at INTEGER NOT NULL, UNIQUE(component_id,version));
CREATE TABLE IF NOT EXISTS dependencies(id TEXT PRIMARY KEY, source_release_id TEXT NOT NULL REFERENCES releases(id) ON DELETE CASCADE, target_component TEXT NOT NULL, constraint_text TEXT NOT NULL, optional INTEGER NOT NULL, platform TEXT NOT NULL, locked INTEGER NOT NULL);
CREATE TABLE IF NOT EXISTS resolve_runs(id TEXT PRIMARY KEY, root_component TEXT NOT NULL, root_constraint TEXT NOT NULL, platform TEXT NOT NULL, allow_prerelease INTEGER NOT NULL, locks TEXT NOT NULL, policy TEXT NOT NULL, input_digest TEXT NOT NULL, status TEXT NOT NULL, error TEXT NOT NULL, created_at INTEGER NOT NULL, finished_at INTEGER NOT NULL DEFAULT 0);
CREATE TABLE IF NOT EXISTS selections(run_id TEXT NOT NULL REFERENCES resolve_runs(id) ON DELETE CASCADE, component_id TEXT NOT NULL, release_id TEXT NOT NULL, version TEXT NOT NULL, reason TEXT NOT NULL, score INTEGER NOT NULL, PRIMARY KEY(run_id,component_id));
CREATE TABLE IF NOT EXISTS edges(run_id TEXT NOT NULL REFERENCES resolve_runs(id) ON DELETE CASCADE, from_component TEXT NOT NULL, to_component TEXT NOT NULL, constraint_text TEXT NOT NULL, optional INTEGER NOT NULL, satisfied INTEGER NOT NULL, path TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS conflicts(id TEXT PRIMARY KEY, run_id TEXT NOT NULL REFERENCES resolve_runs(id) ON DELETE CASCADE, component_id TEXT NOT NULL, kind TEXT NOT NULL, message TEXT NOT NULL, path TEXT NOT NULL, candidates TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS snapshots(id TEXT PRIMARY KEY, run_id TEXT NOT NULL REFERENCES resolve_runs(id), name TEXT NOT NULL, status TEXT NOT NULL, digest TEXT NOT NULL, created_at INTEGER NOT NULL, activated_at INTEGER NOT NULL DEFAULT 0);
CREATE TABLE IF NOT EXISTS snapshot_items(snapshot_id TEXT NOT NULL REFERENCES snapshots(id) ON DELETE CASCADE, component_id TEXT NOT NULL, release_id TEXT NOT NULL, version TEXT NOT NULL, PRIMARY KEY(snapshot_id,component_id));
CREATE TABLE IF NOT EXISTS audit_events(id INTEGER PRIMARY KEY AUTOINCREMENT, aggregate TEXT NOT NULL, aggregate_id TEXT NOT NULL, action TEXT NOT NULL, summary TEXT NOT NULL, created_at INTEGER NOT NULL);
CREATE UNIQUE INDEX IF NOT EXISTS one_active_snapshot ON snapshots(status) WHERE status='active';
CREATE INDEX IF NOT EXISTS idx_releases_component ON releases(component_id);
CREATE INDEX IF NOT EXISTS idx_runs_created ON resolve_runs(created_at DESC);
`
