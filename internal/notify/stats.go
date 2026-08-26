package notify

import (
	"context"
	"sync"
)

// Summary 统计收到的对账通知。
type Summary struct {
	Notifications int
	Missing       int
	Mismatch      int
	Extra         int
}

// CountingSubscriber 累计通知统计。
type CountingSubscriber struct {
	mu      sync.Mutex
	summary Summary
}

// NewCountingSubscriber 创建统计订阅者。
func NewCountingSubscriber() *CountingSubscriber {
	return &CountingSubscriber{}
}

// OnReconciled 累加通知统计。
func (s *CountingSubscriber) OnReconciled(ctx context.Context, msg Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.summary.Notifications++
	s.summary.Missing += msg.Missing
	s.summary.Mismatch += msg.Mismatch
	s.summary.Extra += msg.Extra
	return nil
}

// Snapshot 返回当前统计快照。
func (s *CountingSubscriber) Snapshot() Summary {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.summary
}
