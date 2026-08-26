package diff

import (
	"context"
	"fmt"

	"reconcilesvc/internal/model"
)

// Sink 将对账结果写入差异库，保证失败可重试、重试不重复落库。
type Sink struct {
	store      Store
	dedup      *Dedup
	maxRetries int
}

// NewSink 创建差异写入器。
func NewSink(store Store, maxRetries int) *Sink {
	return &Sink{store: store, dedup: NewDedup(), maxRetries: maxRetries}
}

// WriteResult 写入窗口差异；已提交位点之前与已落库的差异自动跳过。
func (s *Sink) WriteResult(ctx context.Context, window model.Window, result *model.Result, resume ResumeInfo) error {
	for _, entry := range result.Entries {
		key := entry.DedupKey()
		if s.dedup.Seen(key) {
			continue
		}
		if err := s.putWithRetry(ctx, entry); err != nil {
			return fmt.Errorf("write diff %s: %w", key, err)
		}
		s.dedup.Mark(key)
	}
	return nil
}

// putWithRetry 写入失败时按策略重试。
func (s *Sink) putWithRetry(ctx context.Context, entry model.Entry) error {
	var lastErr error
	for attempt := 0; attempt <= s.maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
		}
		if err := s.store.Put(ctx, entry); err != nil {
			lastErr = err
			continue
		}
		return nil
	}
	return lastErr
}
