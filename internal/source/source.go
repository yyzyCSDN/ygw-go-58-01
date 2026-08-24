package source

import (
	"context"

	"reconcilesvc/internal/model"
)

// Reader 提供逐条读取与关闭能力。
type Reader interface {
	Next(ctx context.Context) (*model.Record, error)
	Close() error
	Closed() bool
}

// Source 表示一个可打开读取器的数据源。
type Source interface {
	Open(ctx context.Context) (Reader, error)
	Total() int
	Name() string
}

// MemorySource 是基于内存切片的数据源，借用连接池中的连接。
type MemorySource struct {
	name    string
	records []*model.Record
	pool    *ConnPool
}

// NewMemorySource 创建内存数据源。
func NewMemorySource(name string, records []*model.Record, pool *ConnPool) *MemorySource {
	return &MemorySource{name: name, records: records, pool: pool}
}

// Name 返回数据源名称。
func (s *MemorySource) Name() string { return s.name }

// Total 返回记录总数。
func (s *MemorySource) Total() int { return len(s.records) }

// Open 借用连接并返回读取器。
func (s *MemorySource) Open(ctx context.Context) (Reader, error) {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	return newSliceReader(s.records, conn), nil
}
