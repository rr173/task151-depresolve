package plan

import (
	"sort"

	"depresolve/internal/model"
)

type Action struct {
	ComponentID string `json:"component_id"`
	From        string `json:"from,omitempty"`
	To          string `json:"to"`
	Kind        string `json:"kind"`
}

func Build(previous []model.Selection, current []model.Selection) []Action {
	old := map[string]model.Selection{}
	for _, x := range previous {
		old[x.ComponentID] = x
	}
	out := []Action{}
	for _, x := range current {
		if p, ok := old[x.ComponentID]; !ok {
			out = append(out, Action{ComponentID: x.ComponentID, To: x.Version, Kind: "install"})
		} else if p.Version != x.Version {
			out = append(out, Action{ComponentID: x.ComponentID, From: p.Version, To: x.Version, Kind: "change"})
			delete(old, x.ComponentID)
		} else {
			delete(old, x.ComponentID)
		}
	}
	for _, p := range old {
		out = append(out, Action{ComponentID: p.ComponentID, From: p.Version, Kind: "remove"})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ComponentID < out[j].ComponentID })
	return out
}

func Diff(previous, current []model.Selection) model.SnapshotDiff {
	old := map[string]model.Selection{}
	for _, x := range previous {
		old[x.ComponentID] = x
	}
	d := model.SnapshotDiff{Added: []model.Selection{}, Removed: []model.Selection{}, Changed: []model.VersionChange{}}
	for _, x := range current {
		if p, ok := old[x.ComponentID]; !ok {
			d.Added = append(d.Added, x)
		} else {
			if p.Version != x.Version {
				d.Changed = append(d.Changed, model.VersionChange{ComponentID: x.ComponentID, From: p.Version, To: x.Version})
			}
			delete(old, x.ComponentID)
		}
	}
	for _, x := range old {
		d.Removed = append(d.Removed, x)
	}
	sort.Slice(d.Added, func(i, j int) bool { return d.Added[i].ComponentID < d.Added[j].ComponentID })
	sort.Slice(d.Removed, func(i, j int) bool { return d.Removed[i].ComponentID < d.Removed[j].ComponentID })
	sort.Slice(d.Changed, func(i, j int) bool { return d.Changed[i].ComponentID < d.Changed[j].ComponentID })
	return d
}
