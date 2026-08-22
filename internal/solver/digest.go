package solver

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"

	"depresolve/internal/model"
)

func InputDigest(r model.ResolveRequest) string {
	locks := make([]string, 0, len(r.Locks))
	for k, v := range r.Locks {
		locks = append(locks, k+"="+v)
	}
	sort.Strings(locks)
	canonical := struct {
		Root, Constraint, Platform, Policy string
		Allow                              bool
		Locks                              []string
	}{r.RootComponent, r.RootConstraint, r.Platform, r.Policy, r.AllowPrerelease, locks}
	b, _ := json.Marshal(canonical)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}
func ResultDigest(selections []model.Selection) string {
	cp := append([]model.Selection{}, selections...)
	sort.Slice(cp, func(i, j int) bool { return cp[i].ComponentID < cp[j].ComponentID })
	b, _ := json.Marshal(cp)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}
