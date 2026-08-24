package target

import (
	"context"

	"reconcilesvc/internal/model"
)

// SnapshotService 提供目标快照的版本化访问。
type SnapshotService struct {
	store Store
	log   *UpdateLog
}

// NewSnapshotService 创建快照服务。
func NewSnapshotService(store Store) *SnapshotService {
	return &SnapshotService{store: store}
}

// WithLog 关联版本更新日志。
func (s *SnapshotService) WithLog(log *UpdateLog) *SnapshotService {
	s.log = log
	return s
}

// Latest 返回目标存储当前版本快照。
func (s *SnapshotService) Latest(ctx context.Context) (*model.Snapshot, error) {
	return s.store.Snapshot(ctx)
}

// Version 返回目标存储当前版本号。
func (s *SnapshotService) Version() int64 { return s.store.Version() }

// History 返回目标版本更新历史。
func (s *SnapshotService) History() []VersionRecord {
	if s.log == nil {
		return nil
	}
	return s.log.Entries()
}
