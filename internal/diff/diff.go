package diff

import (
	"context"
	"sync"

	"reconcilesvc/internal/model"
)

// Store 持久化差异条目。
type Store interface {
	Put(ctx context.Context, entry model.Entry) error
	Has(ctx context.Context, dedupKey string) (bool, error)
	List(ctx context.Context, limit int) ([]model.Entry, error)
}

// MemoryStore 是基于内存的差异库。
type MemoryStore struct {
	mu      sync.Mutex
	entries []model.Entry
	seen    map[string]struct{}
	fail    error
}

// NewMemoryStore 创建内存差异库。
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{seen: make(map[string]struct{})}
}

// FailNext 注入一次写入失败，用于故障演练。
func (s *MemoryStore) FailNext(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.fail = err
}

// Put 写入一条差异；重复键自动跳过。
func (s *MemoryStore) Put(ctx context.Context, entry model.Entry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.fail != nil {
		err := s.fail
		s.fail = nil
		return err
	}
	key := entry.DedupKey()
	if _, ok := s.seen[key]; ok {
		return nil
	}
	s.seen[key] = struct{}{}
	s.entries = append(s.entries, entry)
	return nil
}

// Has 判断差异键是否已落库。
func (s *MemoryStore) Has(ctx context.Context, key string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.seen[key]
	return ok, nil
}

// List 返回最近 limit 条差异。
func (s *MemoryStore) List(ctx context.Context, limit int) ([]model.Entry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if limit <= 0 || limit > len(s.entries) {
		limit = len(s.entries)
	}
	out := make([]model.Entry, limit)
	copy(out, s.entries[len(s.entries)-limit:])
	return out, nil
}
