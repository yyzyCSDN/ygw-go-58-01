package notify

import "fmt"

// String 返回通知消息的展示文本。
func (m Message) String() string {
	return fmt.Sprintf(
		"window=%s phase=%s total=%d missing=%d mismatch=%d extra=%d pos=%s",
		m.WindowID, m.Phase, m.Total, m.Missing, m.Mismatch, m.Extra, m.PositionKey,
	)
}
