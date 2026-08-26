package offset

import (
	"sync"

	"reconcilesvc/internal/model"
)

// Journal 记录位点推进历史。
type Journal struct {
	mu      sync.Mutex
	entries []model.Position
}

// NewJournal 创建位点日志。
func NewJournal() *Journal {
	return &Journal{}
}

// Append 追加一次位点快照。
func (j *Journal) Append(pos model.Position) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.entries = append(j.entries, pos)
}

// Entries 返回位点历史的副本。
func (j *Journal) Entries() []model.Position {
	j.mu.Lock()
	defer j.mu.Unlock()
	out := make([]model.Position, len(j.entries))
	copy(out, j.entries)
	return out
}
