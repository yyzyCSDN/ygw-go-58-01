package target

import (
	"context"
	"sync"

	"reconcilesvc/internal/model"
)

// Baseline 表示一次对账使用的目标基线。
type Baseline struct {
	snapshot *model.Snapshot
	version  int64
}

// Version 返回基线版本。
func (b *Baseline) Version() int64 { return b.version }

// Get 返回基线内的记录；记录不存在或为空时返回 nil, false。
func (b *Baseline) Get(key string) (*model.Record, bool) {
	if b.snapshot == nil {
		return nil, false
	}
	rec, ok := b.snapshot.Records[key]
	return rec, ok
}

// Keys 返回基线覆盖的全部键。
func (b *Baseline) Keys() []string {
	if b.snapshot == nil {
		return nil
	}
	return b.snapshot.Keys()
}

// BaselineManager 管理目标基线，保证使用目标存储的最新版本。
type BaselineManager struct {
	store Store
	mu    sync.Mutex
	base  *Baseline
}

// NewBaselineManager 创建基线管理器。
func NewBaselineManager(store Store) *BaselineManager {
	return &BaselineManager{store: store}
}

// Acquire 返回与目标存储版本一致的最新基线；版本变化时自动刷新。
func (m *BaselineManager) Acquire(ctx context.Context) (*Baseline, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.base != nil && m.base.version == m.store.Version() {
		return m.base, nil
	}
	snapshot, err := m.store.Snapshot(ctx)
	if err != nil {
		return nil, err
	}
	m.base = &Baseline{snapshot: snapshot, version: snapshot.Version}
	return m.base, nil
}

// Refresh 强制重新加载目标基线。
func (m *BaselineManager) Refresh(ctx context.Context) (*Baseline, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	snapshot, err := m.store.Snapshot(ctx)
	if err != nil {
		return nil, err
	}
	m.base = &Baseline{snapshot: snapshot, version: snapshot.Version}
	return m.base, nil
}
