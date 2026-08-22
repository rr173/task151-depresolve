package model

type RunStatus string

const (
	RunPending  RunStatus = "pending"
	RunResolved RunStatus = "resolved"
	RunFailed   RunStatus = "failed"
)

type SnapshotStatus string

const (
	SnapshotDraft      SnapshotStatus = "draft"
	SnapshotActive     SnapshotStatus = "active"
	SnapshotRetired    SnapshotStatus = "retired"
	SnapshotRolledBack SnapshotStatus = "rolled_back"
)

type ConflictKind string

const (
	ConflictRange      ConflictKind = "range"
	ConflictPlatform   ConflictKind = "platform"
	ConflictPrerelease ConflictKind = "prerelease"
	ConflictLock       ConflictKind = "lock"
	ConflictCapability ConflictKind = "capability"
	ConflictCycle      ConflictKind = "cycle"
)

func (s SnapshotStatus) CanActivate() bool { return s == SnapshotDraft || s == SnapshotRetired }
