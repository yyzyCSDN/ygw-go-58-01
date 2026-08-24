package notify

import (
	"context"
	"errors"
	"fmt"

	"reconcilesvc/internal/model"
)

// Message 表示一次窗口对账完成通知。
type Message struct {
	WindowID    string
	Phase       model.Phase
	Total       int
	Missing     int
	Mismatch    int
	Extra       int
	PositionKey string
}

// Subscriber 订阅对账结果。
type Subscriber interface {
	OnReconciled(ctx context.Context, msg Message) error
}

// Notifier 向全部订阅者分发通知。
type Notifier struct {
	subscribers []Subscriber
}

// NewNotifier 创建通知器。
func NewNotifier(subscribers ...Subscriber) *Notifier {
	return &Notifier{subscribers: subscribers}
}

// Notify 分发窗口通知；任一订阅者失败时合并上报。
func (n *Notifier) Notify(ctx context.Context, window model.Window, result *model.Result, positionKey string) error {
	msg := Message{
		WindowID:    window.ID,
		Phase:       window.Phase,
		Total:       result.Total,
		Missing:     result.Missing,
		Mismatch:    result.Mismatch,
		Extra:       result.Extra,
		PositionKey: positionKey,
	}
	var errs []error
	for _, sub := range n.subscribers {
		if err := sub.OnReconciled(ctx, msg); err != nil {
			errs = append(errs, fmt.Errorf("%T: %w", sub, err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("notify window %s: %w", window.ID, errors.Join(errs...))
	}
	return nil
}
