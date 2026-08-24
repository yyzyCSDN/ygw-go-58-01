package model

import "fmt"

// String 返回位点的展示文本。
func (p Position) String() string {
	return fmt.Sprintf(
		"%s[%d]@%s(completed=%t)",
		p.Phase, p.WindowIndex, p.Key, p.Completed,
	)
}
