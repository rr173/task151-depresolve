package semver

import "testing"

func TestComparePrerelease(t *testing.T) {
	a, _ := Parse("1.0.0-alpha.1")
	b, _ := Parse("1.0.0")
	if Compare(a, b) >= 0 {
		t.Fatal("release must outrank prerelease")
	}
}
func TestRange(t *testing.T) {
	r, e := ParseRange("^1.2.0")
	if e != nil {
		t.Fatal(e)
	}
	ok, _ := Parse("1.9.0")
	bad, _ := Parse("2.0.0")
	if !r.Allows(ok) || r.Allows(bad) {
		t.Fatal("caret range failed")
	}
}
func TestWildcard(t *testing.T) {
	r, e := ParseRange("1.2.x")
	if e != nil {
		t.Fatal(e)
	}
	v, _ := Parse("1.2.9")
	if !r.Allows(v) {
		t.Fatal("wildcard failed")
	}
}
