package solver

import (
	"fmt"
	"sort"
	"strings"

	"depresolve/internal/catalog"
	"depresolve/internal/graph"
	"depresolve/internal/model"
	"depresolve/internal/semver"
)

type Input struct {
	Request model.ResolveRequest
	Catalog *catalog.Catalog
}
type Result struct {
	Selections []model.Selection
	Edges      []model.Edge
	Conflicts  []model.Conflict
	Graph      *graph.Graph
}
type constraint struct {
	value    string
	source   string
	optional bool
	platform string
	locked   bool
}
type state struct {
	selected    map[string]model.Release
	constraints map[string][]constraint
	edges       []model.Edge
	path        map[string][]string
	conflicts   []model.Conflict
}

func Solve(in Input) (Result, error) {
	if in.Catalog == nil {
		return Result{}, fmt.Errorf("catalog: %w", model.ErrInvalidInput)
	}
	if _, err := in.Catalog.Component(in.Request.RootComponent); err != nil {
		return Result{}, err
	}
	s := &state{selected: map[string]model.Release{}, constraints: map[string][]constraint{}, path: map[string][]string{}}
	s.constraints[in.Request.RootComponent] = []constraint{{value: in.Request.RootConstraint, source: "root", platform: in.Request.Platform}}
	if err := resolveComponent(s, in, in.Request.RootComponent, []string{in.Request.RootComponent}); err != nil {
		return Result{Conflicts: s.conflicts, Graph: buildGraph(s)}, err
	}
	result := Result{Graph: buildGraph(s), Edges: append([]model.Edge{}, s.edges...), Conflicts: append([]model.Conflict{}, s.conflicts...)}
	keys := make([]string, 0, len(s.selected))
	for k := range s.selected {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, component := range keys {
		rel := s.selected[component]
		reason := "highest compatible candidate"
		if lock := in.Request.Locks[component]; lock != "" {
			reason = "pinned by lock"
		}
		result.Selections = append(result.Selections, model.Selection{ComponentID: component, ReleaseID: rel.ID, Version: rel.Version, Reason: reason, Score: score(rel)})
	}
	if len(result.Conflicts) > 0 {
		return result, fmt.Errorf("%w: %s", model.ErrConflict, result.Conflicts[0].Message)
	}
	return result, nil
}

func resolveComponent(s *state, in Input, component string, path []string) error {
	if len(path) > 1 {
		for i := 0; i < len(path)-1; i++ {
			if path[i] == component {
				c := model.Conflict{ComponentID: component, Kind: model.ConflictCycle, Message: "dependency cycle detected", Path: strings.Join(append(path, component), " -> ")}
				s.conflicts = append(s.conflicts, c)
				return model.ErrCycle
			}
		}
	}
	cs := s.constraints[component]
	constraintText := combine(cs)
	candidates, err := in.Catalog.Candidates(component, constraintText, in.Request.Platform, in.Request.AllowPrerelease)
	if err != nil {
		return err
	}
	lock := in.Request.Locks[component]
	if lock != "" {
		var filtered []model.Release
		for _, r := range candidates {
			if r.Version == lock {
				filtered = append(filtered, r)
			}
		}
		candidates = filtered
		if len(candidates) == 0 {
			addConflict(s, component, model.ConflictLock, fmt.Sprintf("lock %s does not satisfy %s", lock, constraintText), path, nil)
			return model.ErrConflict
		}
	}
	if len(candidates) == 0 {
		optional := true
		for _, c := range cs {
			if !c.optional {
				optional = false
				break
			}
		}
		if optional && !isLocked(component, in.Request.Locks) {
			return nil
		}
		kind := model.ConflictRange
		for _, c := range cs {
			if c.platform != "" && c.platform != in.Request.Platform {
				kind = model.ConflictPlatform
			}
		}
		addConflict(s, component, kind, fmt.Sprintf("no candidate for %s on %s", constraintText, in.Request.Platform), path, nil)
		return model.ErrConflict
	}
	for _, candidate := range candidates {
		if violatesCapability(s, in, candidate, component, path) {
			continue
		}
		snapshot := cloneState(s)
		s.selected[component] = candidate
		s.path[component] = append([]string{}, path...)
		failed := false
		for _, dep := range candidate.Dependencies {
			if dep.Platform != "" && dep.Platform != in.Request.Platform {
				if dep.Optional && !dep.Locked {
					continue
				}
				addConflict(s, component, model.ConflictPlatform, fmt.Sprintf("dependency %s requires platform %s", dep.TargetComponent, dep.Platform), path, nil)
				failed = true
				break
			}
			s.constraints[dep.TargetComponent] = append(s.constraints[dep.TargetComponent], constraint{value: dep.Constraint, source: component + "@" + candidate.Version, optional: dep.Optional, platform: dep.Platform, locked: dep.Locked})
			edge := model.Edge{FromComponent: component, ToComponent: dep.TargetComponent, Constraint: dep.Constraint, Optional: dep.Optional, Satisfied: true, Path: strings.Join(append(path, dep.TargetComponent), " -> ")}
			s.edges = append(s.edges, edge)
			if err := resolveComponent(s, in, dep.TargetComponent, append(path, dep.TargetComponent)); err != nil {
				failed = true
				break
			}
		}
		if !failed {
			return nil
		}
		*s = *snapshot
	}
	addConflict(s, component, model.ConflictRange, fmt.Sprintf("all candidates for %s lead to a conflict", component), path, nil)
	return model.ErrConflict
}

func cloneState(s *state) *state {
	n := &state{selected: map[string]model.Release{}, constraints: map[string][]constraint{}, edges: append([]model.Edge{}, s.edges...), path: map[string][]string{}, conflicts: append([]model.Conflict{}, s.conflicts...)}
	for k, v := range s.selected {
		n.selected[k] = v
	}
	for k, v := range s.constraints {
		n.constraints[k] = append([]constraint{}, v...)
	}
	for k, v := range s.path {
		n.path[k] = append([]string{}, v...)
	}
	return n
}
func combine(cs []constraint) string {
	if len(cs) == 0 {
		return "*"
	}
	parts := make([]string, 0, len(cs))
	for _, c := range cs {
		if c.value == "" {
			parts = append(parts, "*")
		} else {
			parts = append(parts, c.value)
		}
	}
	return strings.Join(parts, " ")
}
func isLocked(id string, locks map[string]string) bool { _, ok := locks[id]; return ok }
func addConflict(s *state, component string, kind model.ConflictKind, msg string, path []string, candidates []string) {
	s.conflicts = append(s.conflicts, model.Conflict{ComponentID: component, Kind: kind, Message: msg, Path: strings.Join(path, " -> "), Candidates: candidates})
}
func violatesCapability(s *state, in Input, r model.Release, component string, path []string) bool {
	for _, other := range s.selected {
		for _, cap := range other.Capabilities {
			for _, blocked := range r.Conflicts {
				if cap == blocked {
					addConflict(s, component, model.ConflictCapability, fmt.Sprintf("%s conflicts with capability %s", component, cap), path, nil)
					return true
				}
			}
		}
		for _, cap := range r.Capabilities {
			for _, blocked := range other.Conflicts {
				if cap == blocked {
					addConflict(s, component, model.ConflictCapability, fmt.Sprintf("capability %s conflicts with selected release", cap), path, nil)
					return true
				}
			}
		}
	}
	return false
}
func score(r model.Release) int {
	v, _ := semver.Parse(r.Version)
	return v.Major*1000000 + v.Minor*1000 + v.Patch
}
func buildGraph(s *state) *graph.Graph {
	g := graph.New()
	for c, r := range s.selected {
		g.AddNode(c, r)
	}
	for _, e := range s.edges {
		g.AddEdge(graph.Edge{From: e.FromComponent, To: e.ToComponent, Constraint: e.Constraint, Optional: e.Optional, Satisfied: e.Satisfied})
	}
	return g
}
