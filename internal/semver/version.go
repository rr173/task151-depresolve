package semver

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var versionPattern = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-([0-9A-Za-z.-]+))?(?:\+[0-9A-Za-z.-]+)?$`)

type Version struct {
	Major int
	Minor int
	Patch int
	Pre   []string
	Raw   string
}

func Parse(raw string) (Version, error) {
	raw = strings.TrimSpace(raw)
	m := versionPattern.FindStringSubmatch(raw)
	if m == nil {
		return Version{}, fmt.Errorf("invalid version %q", raw)
	}
	major, _ := strconv.Atoi(m[1])
	minor, _ := strconv.Atoi(m[2])
	patch, _ := strconv.Atoi(m[3])
	v := Version{Major: major, Minor: minor, Patch: patch, Raw: raw}
	if m[4] != "" {
		v.Pre = strings.Split(m[4], ".")
	}
	return v, nil
}

func (v Version) String() string {
	if v.Raw != "" {
		return v.Raw
	}
	base := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if len(v.Pre) > 0 {
		return base + "-" + strings.Join(v.Pre, ".")
	}
	return base
}
func (v Version) IsPrerelease() bool { return len(v.Pre) > 0 }

func Compare(a, b Version) int {
	if a.Major != b.Major {
		if a.Major < b.Major {
			return -1
		}
		return 1
	}
	if a.Minor != b.Minor {
		if a.Minor < b.Minor {
			return -1
		}
		return 1
	}
	if a.Patch != b.Patch {
		if a.Patch < b.Patch {
			return -1
		}
		return 1
	}
	if len(a.Pre) == 0 && len(b.Pre) == 0 {
		return 0
	}
	if len(a.Pre) == 0 {
		return 1
	}
	if len(b.Pre) == 0 {
		return -1
	}
	for i := 0; i < len(a.Pre) && i < len(b.Pre); i++ {
		x, y := a.Pre[i], b.Pre[i]
		if x == y {
			continue
		}
		xn, xe := strconv.Atoi(x)
		yn, ye := strconv.Atoi(y)
		if xe == nil && ye == nil {
			if xn < yn {
				return -1
			}
			return 1
		}
		if xe == nil {
			return -1
		}
		if ye == nil {
			return 1
		}
		if x < y {
			return -1
		}
		return 1
	}
	if len(a.Pre) < len(b.Pre) {
		return -1
	}
	if len(a.Pre) > len(b.Pre) {
		return 1
	}
	return 0
}

func SortDescending(vs []Version) {
	for i := 1; i < len(vs); i++ {
		x := vs[i]
		j := i - 1
		for j >= 0 && Compare(vs[j], x) < 0 {
			vs[j+1] = vs[j]
			j--
		}
		vs[j+1] = x
	}
}
