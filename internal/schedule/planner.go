package schedule

import (
	"fmt"
	"sort"

	"reconcilesvc/internal/model"
)

// Planner 将键集合按排序顺序切分为连续窗口。
type Planner struct {
	size int
}

// NewPlanner 创建窗口规划器。
func NewPlanner(windowSize int) *Planner {
	if windowSize <= 0 {
		windowSize = 1
	}
	return &Planner{size: windowSize}
}

// Plan 生成覆盖全部键的连续窗口。
func (p *Planner) Plan(keys []string, phase model.Phase) []model.Window {
	if len(keys) == 0 {
		return nil
	}
	sorted := append([]string(nil), keys...)
	sort.Strings(sorted)
	total := len(sorted)
	count := (total + p.size - 1) / p.size
	windows := make([]model.Window, 0, count)
	for i := 0; i < count-1; i++ {
		start := i * p.size
		end := start + p.size
		if end > total {
			end = total
		}
		windows = append(windows, model.Window{
			ID:    fmt.Sprintf("%s-w%02d", phase, i),
			Index: i,
			Total: count,
			Start: sorted[start],
			End:   sorted[end-1],
			Phase: phase,
		})
	}
	return windows
}

// ResumeWindows 返回位点之后仍需对账的窗口。
func (p *Planner) ResumeWindows(keys []string, phase model.Phase, resumeKey string) []model.Window {
	windows := p.Plan(keys, phase)
	if resumeKey == "" {
		return windows
	}
	kept := make([]model.Window, 0, len(windows))
	for _, w := range windows {
		if w.End > resumeKey {
			kept = append(kept, w)
		}
	}
	return kept
}
