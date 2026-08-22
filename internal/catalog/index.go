package catalog

import (
	"sort"

	"depresolve/internal/model"
)

type Index struct {
	ByComponent  map[string][]model.Release
	ByCapability map[string][]string
	Platforms    map[string]map[string]bool
}

func BuildIndex(c *Catalog) Index {
	idx := Index{ByComponent: map[string][]model.Release{}, ByCapability: map[string][]string{}, Platforms: map[string]map[string]bool{}}
	for id, rels := range c.Releases {
		idx.ByComponent[id] = append([]model.Release{}, rels...)
		for _, r := range rels {
			if _, ok := idx.Platforms[id]; !ok {
				idx.Platforms[id] = map[string]bool{}
			}
			for _, p := range r.Platforms {
				idx.Platforms[id][p] = true
			}
			for _, cap := range r.Capabilities {
				idx.ByCapability[cap] = append(idx.ByCapability[cap], r.ID)
			}
		}
	}
	for k := range idx.ByCapability {
		sort.Strings(idx.ByCapability[k])
	}
	return idx
}
