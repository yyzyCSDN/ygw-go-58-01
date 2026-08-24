package compare

import (
	"fmt"

	"reconcilesvc/internal/model"
)

// Merge 将 src 的结果聚合进 dst。
func Merge(dst, src *model.Result) {
	if dst == nil || src == nil {
		return
	}
	dst.Total += src.Total
	dst.Missing += src.Missing
	dst.Mismatch += src.Mismatch
	dst.Extra += src.Extra
	dst.Entries = append(dst.Entries, src.Entries...)
	if dst.StartedAt.IsZero() {
		dst.StartedAt = src.StartedAt
	}
	dst.FinishedAt = src.FinishedAt
}

// Describe 返回结果的可读摘要。
func Describe(result *model.Result) string {
	if result == nil {
		return "no result"
	}
	return fmt.Sprintf(
		"total=%d missing=%d mismatch=%d extra=%d",
		result.Total, result.Missing, result.Mismatch, result.Extra,
	)
}
