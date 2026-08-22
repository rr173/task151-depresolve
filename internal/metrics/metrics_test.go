package metrics

import "testing"

func TestCounters(t *testing.T) {
	c := &Counters{}
	c.IncRuns()
	c.IncResolved()
	if x := c.Snapshot(); x.Runs != 1 || x.Resolved != 1 {
		t.Fatalf("%+v", x)
	}
}
