package offset

import (
	"fmt"

	"reconcilesvc/internal/model"
)

// ValidateProgress 校验位点历史单调不回退。
func ValidateProgress(positions []model.Position) error {
	for i := 1; i < len(positions); i++ {
		prev, cur := positions[i-1], positions[i]
		if cur.WindowIndex < prev.WindowIndex {
			return fmt.Errorf(
				"progress regressed at %d: window %d -> %d",
				i, prev.WindowIndex, cur.WindowIndex,
			)
		}
	}
	return nil
}
