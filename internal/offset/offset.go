package offset

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"reconcilesvc/internal/diff"
	"reconcilesvc/internal/model"
)

// Store 保存对账位点。
type Store interface {
	Save(ctx context.Context, pos model.Position) error
	Load(ctx context.Context) (model.Position, error)
}

// MemoryStore 是基于内存的位点存储。
type MemoryStore struct {
	mu  sync.Mutex
	pos model.Position
}

// NewMemoryStore 创建位点存储。
func NewMemoryStore() *MemoryStore { return &MemoryStore{} }

// Save 保存位点。
func (s *MemoryStore) Save(ctx context.Context, pos model.Position) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pos = pos
	return nil
}

// Load 读取位点。
func (s *MemoryStore) Load(ctx context.Context) (model.Position, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.pos, nil
}

// Manager 管理对账位点。
type Manager struct {
	store   Store
	pos     model.Position
	journal *Journal
	mu      sync.Mutex
}

// NewManager 创建位点管理器。
func NewManager(store Store) *Manager { return NewManagerWithJournal(store, nil) }

// NewManagerWithJournal 创建带推进历史的位点管理器。
func NewManagerWithJournal(store Store, journal *Journal) *Manager {
	return &Manager{store: store, journal: journal}
}

// Position 返回当前位点。
func (m *Manager) Position() model.Position {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.pos
}

// Advance 在窗口对账完成后推进位点并持久化。
func (m *Manager) Advance(ctx context.Context, window model.Window) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if window.Index < m.pos.WindowIndex {
		return fmt.Errorf("window %d older than position %d", window.Index, m.pos.WindowIndex)
	}
	next := model.Position{
		Phase:       window.Phase,
		WindowIndex: window.Index,
		Key:         window.End,
	}
	if err := m.store.Save(ctx, next); err != nil {
		return fmt.Errorf("persist offset: %w", err)
	}
	m.pos = next
	if m.journal != nil {
		m.journal.Append(next)
	}
	return nil
}

// Complete 标记本轮对账完成。
func (m *Manager) Complete(ctx context.Context, window model.Window) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	next := m.pos
	next.Completed = true
	if window.End != "" {
		next.Key = window.End
		next.Phase = window.Phase
	}
	if err := m.store.Save(ctx, next); err != nil {
		return fmt.Errorf("persist completed offset: %w", err)
	}
	m.pos = next
	if m.journal != nil {
		m.journal.Append(next)
	}
	return nil
}

// ResumeInfo 返回已提交位点信息，供差异写入去重。
func (m *Manager) ResumeInfo(ctx context.Context) (diff.ResumeInfo, error) {
	pos, err := m.store.Load(ctx)
	if err != nil {
		return diff.ResumeInfo{}, err
	}
	return diff.ResumeInfo{CommittedKey: pos.Key, Phase: pos.Phase}, nil
}

// Continuity 校验窗口序列连续覆盖全部键。
func Continuity(windows []model.Window, keys []string) error {
	sorted := append([]string(nil), keys...)
	sort.Strings(sorted)
	covered := 0
	for _, key := range sorted {
		for _, w := range windows {
			if w.Contains(key) {
				covered++
				break
			}
		}
	}
	if covered != len(sorted) {
		return fmt.Errorf("windows cover %d of %d keys", covered, len(sorted))
	}
	return nil
}
