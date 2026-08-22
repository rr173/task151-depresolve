package model

type CreateComponentRequest struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Ecosystem  string `json:"ecosystem"`
	StableOnly *bool  `json:"stable_only"`
}
type CreateReleaseRequest struct {
	ID           string            `json:"id"`
	Version      string            `json:"version"`
	Platforms    []string          `json:"platforms"`
	Capabilities []string          `json:"capabilities"`
	Conflicts    []string          `json:"conflicts"`
	Dependencies []DependencyInput `json:"dependencies"`
}
type DependencyInput struct {
	ID              string `json:"id"`
	TargetComponent string `json:"target_component"`
	Constraint      string `json:"constraint"`
	Optional        bool   `json:"optional"`
	Platform        string `json:"platform"`
	Locked          bool   `json:"locked"`
}
type ResolveResponse struct {
	Run        ResolveRun  `json:"run"`
	Selections []Selection `json:"selections"`
	Edges      []Edge      `json:"edges"`
	Conflicts  []Conflict  `json:"conflicts"`
}
type CreateSnapshotRequest struct {
	Name string `json:"name"`
}
type Page[T any] struct {
	Items  []T `json:"items"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
}

func (r CreateComponentRequest) Normalize() Component {
	stable := true
	if r.StableOnly != nil {
		stable = *r.StableOnly
	}
	return Component{ID: r.ID, Name: r.Name, Ecosystem: r.Ecosystem, StableOnly: stable}
}
