package model

import "fmt"

// Phase 表示对账阶段。
type Phase int

const (
	PhaseFull Phase = iota
	PhaseIncremental
)

// String 返回阶段的稳定名称。
func (p Phase) String() string {
	switch p {
	case PhaseFull:
		return "full"
	case PhaseIncremental:
		return "incremental"
	default:
		return "unknown"
	}
}

// Window 表示对账窗口，覆盖 [Start, End] 的键区间。
type Window struct {
	ID    string
	Index int
	Total int
	Start string
	End   string
	Phase Phase
}

// Contains 判断键是否属于窗口区间。
func (w Window) Contains(key string) bool {
	return key >= w.Start && key <= w.End
}

// String 返回窗口的展示名。
func (w Window) String() string {
	return fmt.Sprintf("%s[%s..%s]", w.ID, w.Start, w.End)
}
