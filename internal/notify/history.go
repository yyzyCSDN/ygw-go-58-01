package notify

import (
	"context"
	"sync"
)

// History 保存最近的通知消息。
type History struct {
	mu       sync.Mutex
	max      int
	messages []Message
}

// NewHistory 创建消息历史。
func NewHistory(max int) *History {
	if max <= 0 {
		max = 32
	}
	return &History{max: max}
}

// Append 追加一条消息，超出容量时丢弃最旧消息。
func (h *History) Append(msg Message) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.messages = append(h.messages, msg)
	if len(h.messages) > h.max {
		h.messages = h.messages[len(h.messages)-h.max:]
	}
}

// Messages 返回消息历史的副本。
func (h *History) Messages() []Message {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]Message, len(h.messages))
	copy(out, h.messages)
	return out
}

// historySubscriber 将通知写入历史。
type historySubscriber struct {
	history *History
}

// NewHistorySubscriber 创建写入历史的订阅者。
func NewHistorySubscriber(history *History) Subscriber {
	return &historySubscriber{history: history}
}

// OnReconciled 将消息追加进历史。
func (h *historySubscriber) OnReconciled(ctx context.Context, msg Message) error {
	h.history.Append(msg)
	return nil
}
