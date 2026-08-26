package target

import (
	"context"

	"reconcilesvc/internal/model"
)

// Store 提供目标侧数据快照。
type Store interface {
	Snapshot(ctx context.Context) (*model.Snapshot, error)
	Version() int64
}

// MemoryStore 是基于内存的目标存储。
type MemoryStore struct {
	version int64
	records map[string]*model.Record
}

// NewMemoryStore 创建目标存储。
func NewMemoryStore(version int64, records map[string]*model.Record) *MemoryStore {
	return &MemoryStore{version: version, records: records}
}

// Snapshot 返回当前版本快照。
func (s *MemoryStore) Snapshot(ctx context.Context) (*model.Snapshot, error) {
	return &model.Snapshot{Version: s.version, Records: s.records}, nil
}

// Version 返回当前版本号。
func (s *MemoryStore) Version() int64 { return s.version }

// Update 用新版本数据替换存储。
func (s *MemoryStore) Update(version int64, records map[string]*model.Record) {
	s.version = version
	s.records = records
}
