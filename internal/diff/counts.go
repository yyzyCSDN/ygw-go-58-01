package diff

import "reconcilesvc/internal/model"

// Counts 统计差异列表中各类型的数量。
func Counts(entries []model.Entry) map[string]int {
	counts := map[string]int{"missing": 0, "mismatch": 0, "extra": 0}
	for _, entry := range entries {
		counts[entry.Kind.String()]++
	}
	return counts
}
