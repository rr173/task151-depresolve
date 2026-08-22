package model

import "time"

type Component struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Ecosystem  string    `json:"ecosystem"`
	StableOnly bool      `json:"stable_only"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Dependency struct {
	ID              string `json:"id"`
	SourceReleaseID string `json:"source_release_id"`
	TargetComponent string `json:"target_component"`
	Constraint      string `json:"constraint"`
	Optional        bool   `json:"optional"`
	Platform        string `json:"platform"`
	Locked          bool   `json:"locked"`
}

type Release struct {
	ID           string       `json:"id"`
	ComponentID  string       `json:"component_id"`
	Version      string       `json:"version"`
	Platforms    []string     `json:"platforms"`
	Capabilities []string     `json:"capabilities"`
	Conflicts    []string     `json:"conflicts"`
	Dependencies []Dependency `json:"dependencies"`
	Published    bool         `json:"published"`
	CreatedAt    time.Time    `json:"created_at"`
}

type ResolveRequest struct {
	ID              string            `json:"id"`
	RootComponent   string            `json:"root_component"`
	RootConstraint  string            `json:"root_constraint"`
	Platform        string            `json:"platform"`
	AllowPrerelease bool              `json:"allow_prerelease"`
	Locks           map[string]string `json:"locks"`
	Policy          string            `json:"policy"`
}

type ResolveRun struct {
	ID              string            `json:"id"`
	RootComponent   string            `json:"root_component"`
	RootConstraint  string            `json:"root_constraint"`
	Platform        string            `json:"platform"`
	AllowPrerelease bool              `json:"allow_prerelease"`
	Locks           map[string]string `json:"locks"`
	Policy          string            `json:"policy"`
	InputDigest     string            `json:"input_digest"`
	Status          RunStatus         `json:"status"`
	Error           string            `json:"error,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	FinishedAt      *time.Time        `json:"finished_at,omitempty"`
}

type Selection struct {
	RunID       string `json:"run_id"`
	ComponentID string `json:"component_id"`
	ReleaseID   string `json:"release_id"`
	Version     string `json:"version"`
	Reason      string `json:"reason"`
	Score       int    `json:"score"`
}

type Edge struct {
	RunID         string `json:"run_id"`
	FromComponent string `json:"from_component"`
	ToComponent   string `json:"to_component"`
	Constraint    string `json:"constraint"`
	Optional      bool   `json:"optional"`
	Satisfied     bool   `json:"resolved"`
	Path          string `json:"path"`
}

type Conflict struct {
	ID          string       `json:"id"`
	RunID       string       `json:"run_id"`
	ComponentID string       `json:"component_id"`
	Kind        ConflictKind `json:"kind"`
	Message     string       `json:"message"`
	Path        string       `json:"path"`
	Candidates  []string     `json:"candidates"`
}

type Snapshot struct {
	ID          string         `json:"id"`
	RunID       string         `json:"run_id"`
	Name        string         `json:"name"`
	Status      SnapshotStatus `json:"status"`
	Digest      string         `json:"digest"`
	CreatedAt   time.Time      `json:"created_at"`
	ActivatedAt *time.Time     `json:"activated_at,omitempty"`
}

type SnapshotItem struct {
	SnapshotID  string `json:"snapshot_id"`
	ComponentID string `json:"component_id"`
	ReleaseID   string `json:"release_id"`
	Version     string `json:"version"`
}
type AuditEvent struct {
	ID          int64     `json:"id"`
	Aggregate   string    `json:"aggregate"`
	AggregateID string    `json:"aggregate_id"`
	Action      string    `json:"action"`
	Summary     string    `json:"summary"`
	CreatedAt   time.Time `json:"created_at"`
}
type Stats struct {
	Components      int `json:"components"`
	Releases        int `json:"releases"`
	Resolved        int `json:"resolved"`
	Conflicts       int `json:"conflicts"`
	ActiveSnapshots int `json:"active_snapshots"`
}
type SnapshotDiff struct {
	Added   []Selection     `json:"added"`
	Removed []Selection     `json:"removed"`
	Changed []VersionChange `json:"changed"`
}
type VersionChange struct {
	ComponentID string `json:"component_id"`
	From        string `json:"from"`
	To          string `json:"to"`
}
