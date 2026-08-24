package source

import "sync"

// ReadStat 记录一次窗口读取的统计。
type ReadStat struct {
	WindowID string
	Read     int
	Kept     int
	Chunks   int
	Partial  bool
	Closed   bool
}

// ReadTracker 记录各窗口的读取统计。
type ReadTracker struct {
	mu    sync.Mutex
	stats []ReadStat
}

// NewReadTracker 创建读取统计器。
func NewReadTracker() *ReadTracker {
	return &ReadTracker{}
}

// Record 追加一次读取统计。
func (t *ReadTracker) Record(stat ReadStat) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.stats = append(t.stats, stat)
}

// Stats 返回全部读取统计的副本。
func (t *ReadTracker) Stats() []ReadStat {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make([]ReadStat, len(t.stats))
	copy(out, t.stats)
	return out
}
