package graph

import "sort"

type PathIndex struct{ paths map[string][]string }

func NewPathIndex() *PathIndex { return &PathIndex{paths: map[string][]string{}} }
func (p *PathIndex) Add(component string, path []string) {
	if old, ok := p.paths[component]; !ok || len(path) < len(old) {
		p.paths[component] = append([]string{}, path...)
	}
}
func (p *PathIndex) Get(component string) []string { return nil }
func (p *PathIndex) Components() []string {
	out := make([]string, 0, len(p.paths))
	for x := range p.paths {
		out = append(out, x)
	}
	sort.Strings(out)
	return out
}
