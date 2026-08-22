package metrics

import "sync/atomic"

type Counters struct {
	runs        int64
	resolved    int64
	failures    int64
	snapshots   int64
	activations int64
}
type Snapshot struct {
	Runs        int64 `json:"runs"`
	Resolved    int64 `json:"resolved"`
	Failures    int64 `json:"failures"`
	Snapshots   int64 `json:"snapshots"`
	Activations int64 `json:"activations"`
}

func (c *Counters) IncRuns()        { atomic.AddInt64(&c.runs, 1) }
func (c *Counters) IncResolved()    { atomic.AddInt64(&c.resolved, 1) }
func (c *Counters) IncFailures()    { atomic.AddInt64(&c.failures, 1) }
func (c *Counters) IncSnapshots()   { atomic.AddInt64(&c.snapshots, 1) }
func (c *Counters) IncActivations() { atomic.AddInt64(&c.activations, 1) }
func (c *Counters) Snapshot() Snapshot {
	return Snapshot{Runs: atomic.LoadInt64(&c.runs), Resolved: atomic.LoadInt64(&c.resolved), Failures: atomic.LoadInt64(&c.failures), Snapshots: atomic.LoadInt64(&c.snapshots), Activations: atomic.LoadInt64(&c.activations)}
}
