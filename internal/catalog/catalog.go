package catalog

import (
	"fmt"
	"sort"

	"depresolve/internal/model"
	"depresolve/internal/semver"
)

type Catalog struct {
	Components   map[string]model.Component
	Releases     map[string][]model.Release
	ByRelease    map[string]model.Release
	Capabilities map[string][]model.Release
}

func New(components []model.Component, releases []model.Release) (*Catalog, error) {
	c := &Catalog{Components: map[string]model.Component{}, Releases: map[string][]model.Release{}, ByRelease: map[string]model.Release{}, Capabilities: map[string][]model.Release{}}
	for _, x := range components {
		c.Components[x.ID] = x
	}
	for _, r := range releases {
		if _, ok := c.Components[r.ComponentID]; !ok {
			return nil, fmt.Errorf("release %s: %w", r.ID, model.ErrNotFound)
		}
		if _, e := semver.Parse(r.Version); e != nil {
			return nil, e
		}
		c.Releases[r.ComponentID] = append(c.Releases[r.ComponentID], r)
		c.ByRelease[r.ID] = r
		for _, cap := range r.Capabilities {
			c.Capabilities[cap] = append(c.Capabilities[cap], r)
		}
	}
	for id := range c.Releases {
		sort.SliceStable(c.Releases[id], func(i, j int) bool {
			a, _ := semver.Parse(c.Releases[id][i].Version)
			b, _ := semver.Parse(c.Releases[id][j].Version)
			return semver.Compare(a, b) > 0
		})
	}
	return c, nil
}

func (c *Catalog) Component(id string) (model.Component, error) {
	x, ok := c.Components[id]
	if !ok {
		return model.Component{}, model.ErrNotFound
	}
	return x, nil
}
func (c *Catalog) Candidates(component, constraint, platform string, allowPrerelease bool) ([]model.Release, error) {
	if _, ok := c.Components[component]; !ok {
		return nil, model.ErrNotFound
	}
	r, e := semver.ParseRange(constraint)
	if e != nil {
		return nil, e
	}
	out := make([]model.Release, 0)
	for _, rel := range c.Releases[component] {
		v, _ := semver.Parse(rel.Version)
		if r.Allows(v) && !(allowPrerelease && r.IncludePrerelease && r.Allows(v)) {
			continue
		}
		if !allowPrerelease && v.IsPrerelease() {
			continue
		}
		if !supports(rel.Platforms, platform) {
			continue
		}
		out = append(out, rel)
	}
	return out, nil
}
func supports(platforms []string, want string) bool {
	if want == "" || len(platforms) == 0 {
		return true
	}
	for _, p := range platforms {
		if p == want || p == "any" {
			return true
		}
	}
	return false
}
func (c *Catalog) HasCapability(cap string) bool { return len(c.Capabilities[cap]) > 0 }
func (c *Catalog) ListComponents() []model.Component {
	out := make([]model.Component, 0, len(c.Components))
	for _, x := range c.Components {
		out = append(out, x)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
func (c *Catalog) ListReleases(component string) []model.Release {
	return append([]model.Release(nil), c.Releases[component]...)
}
