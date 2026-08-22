package plan

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"depresolve/internal/model"
	"sort"
)

func Digest(items []model.SnapshotItem) string {
	cp := append([]model.SnapshotItem{}, items...)
	sort.Slice(cp, func(i, j int) bool { return cp[i].ComponentID < cp[j].ComponentID })
	b, _ := json.Marshal(cp)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}
func SelectionMap(items []model.SnapshotItem) []model.Selection {
	out := make([]model.Selection, 0, len(items))
	for _, x := range items {
		out = append(out, model.Selection{ComponentID: x.ComponentID, ReleaseID: x.ReleaseID, Version: x.Version})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ComponentID < out[j].ComponentID })
	return out
}
