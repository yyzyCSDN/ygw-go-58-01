package notify

import "context"

// ChannelSubscriber 将通知写入缓冲 channel。
type ChannelSubscriber struct {
	ch chan Message
}

// NewChannelSubscriber 创建通道订阅者。
func NewChannelSubscriber(buffer int) *ChannelSubscriber {
	return &ChannelSubscriber{ch: make(chan Message, buffer)}
}

// OnReconciled 向 channel 发送通知。
func (s *ChannelSubscriber) OnReconciled(ctx context.Context, msg Message) error {
	select {
	case s.ch <- msg:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Chan 返回只读通知通道。
func (s *ChannelSubscriber) Chan() <-chan Message { return s.ch }
