package target

import (
	"sync"
)

// VersionRecord 记录一次目标版本更新。
type VersionRecord struct {
	Version int64
	Keys    int
}

// UpdateLog 记录目标版本更新历史。
type UpdateLog struct {
	mu      sync.Mutex
	entries []VersionRecord
}

// NewUpdateLog 创建版本日志。
func NewUpdateLog() *UpdateLog {
	return &UpdateLog{}
}

// Append 追加一次版本更新记录。
func (l *UpdateLog) Append(record VersionRecord) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = append(l.entries, record)
}

// Entries 返回版本更新历史的副本。
func (l *UpdateLog) Entries() []VersionRecord {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]VersionRecord, len(l.entries))
	copy(out, l.entries)
	return out
}
