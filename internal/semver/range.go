package semver

import (
	"fmt"
	"strings"
)

type Comparator struct {
	Op      string
	Version Version
}
type Range struct {
	Any               bool
	Comparators       []Comparator
	IncludePrerelease bool
}

func ParseRange(raw string) (Range, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "*" || strings.EqualFold(raw, "latest") {
		return Range{Any: true}, nil
	}
	if strings.HasPrefix(raw, "^") {
		v, e := Parse(strings.TrimPrefix(raw, "^"))
		if e != nil {
			return Range{}, e
		}
		upper := Version{Major: v.Major + 1, Raw: fmt.Sprintf("%d.0.0", v.Major+1)}
		return Range{Comparators: []Comparator{{Op: ">=", Version: v}, {Op: "<", Version: upper}}, IncludePrerelease: v.IsPrerelease()}, nil
	}
	if strings.HasPrefix(raw, "~") {
		v, e := Parse(strings.TrimPrefix(raw, "~"))
		if e != nil {
			return Range{}, e
		}
		upper := Version{Major: v.Major, Minor: v.Minor + 1, Raw: fmt.Sprintf("%d.%d.0", v.Major, v.Minor+1)}
		return Range{Comparators: []Comparator{{Op: ">=", Version: v}, {Op: "<", Version: upper}}, IncludePrerelease: v.IsPrerelease()}, nil
	}
	parts := strings.Fields(raw)
	out := Range{}
	for _, part := range parts {
		op := "="
		value := part
		for _, candidate := range []string{"<=", ">=", "<", ">", "="} {
			if strings.HasPrefix(part, candidate) {
				op = candidate
				value = strings.TrimPrefix(part, candidate)
				break
			}
		}
		if strings.Contains(value, "x") || strings.Contains(value, "X") || strings.Contains(value, "*") {
			c, e := wildcard(value)
			if e != nil {
				return Range{}, e
			}
			out.Comparators = append(out.Comparators, c...)
			continue
		}
		v, e := Parse(value)
		if e != nil {
			return Range{}, e
		}
		out.Comparators = append(out.Comparators, Comparator{Op: op, Version: v})
		if v.IsPrerelease() {
			out.IncludePrerelease = true
		}
	}
	return out, nil
}

func wildcard(raw string) ([]Comparator, error) {
	p := strings.Split(strings.ReplaceAll(strings.ReplaceAll(raw, "X", "x"), "*", "x"), ".")
	if len(p) > 3 {
		return nil, fmt.Errorf("invalid wildcard %q", raw)
	}
	for len(p) < 3 {
		p = append(p, "x")
	}
	nums := make([]int, 2)
	for i := 0; i < 2; i++ {
		if p[i] == "x" {
			break
		}
		var e error
		nums[i], e = atoi(p[i])
		if e != nil {
			return nil, e
		}
	}
	if p[0] == "x" {
		return nil, nil
	}
	lower := Version{Major: nums[0], Raw: fmt.Sprintf("%d.0.0", nums[0])}
	upper := Version{Major: nums[0] + 1, Raw: fmt.Sprintf("%d.0.0", nums[0]+1)}
	if p[1] != "x" {
		lower.Minor = nums[1]
		lower.Raw = fmt.Sprintf("%d.%d.0", nums[0], nums[1])
		upper = Version{Major: nums[0], Minor: nums[1] + 1, Raw: fmt.Sprintf("%d.%d.0", nums[0], nums[1]+1)}
	}
	return []Comparator{{Op: ">=", Version: lower}, {Op: "<", Version: upper}}, nil
}
func atoi(s string) (int, error) {
	var n int
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, fmt.Errorf("invalid version component %q", s)
		}
		n = n*10 + int(r-'0')
	}
	return n, nil
}

func (r Range) Allows(v Version) bool {
	for _, c := range r.Comparators {
		cmp := Compare(v, c.Version)
		ok := false
		switch c.Op {
		case "=":
			ok = cmp == 0
		case ">":
			ok = cmp > 0
		case ">=":
			ok = cmp >= 0
		case "<":
			ok = cmp < 0
		case "<=":
			ok = cmp <= 0
		}
		if !ok {
			return false
		}
	}
	return true
}
func (r Range) String() string {
	if r.Any {
		return "*"
	}
	out := make([]string, 0, len(r.Comparators))
	for _, c := range r.Comparators {
		out = append(out, c.Op+c.Version.String())
	}
	return strings.Join(out, " ")
}

func Intersect(a, b Range) Range {
	if a.Any {
		return b
	}
	if b.Any {
		return a
	}
	out := Range{Comparators: append(append([]Comparator{}, a.Comparators...), b.Comparators...), IncludePrerelease: a.IncludePrerelease || b.IncludePrerelease}
	return out
}
