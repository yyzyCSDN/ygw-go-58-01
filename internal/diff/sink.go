package diff

import (
	"context"
	"fmt"
	"time"

	"reconcilesvc/internal/model"
)

// retryBackoff 是两次写入重试之间的固定退避间隔。
const retryBackoff = 10 * time.Millisecond

// Sink 将对账结果写入差异库，保证失败可重试、重试不重复落库。
type Sink struct {
	store      Store
	dedup      *Dedup
	maxRetries int
}

// NewSink 创建差异写入器。maxRetries 为每条差异写入失败后的最大重试次数，小于 1 时按 1 处理。
func NewSink(store Store, maxRetries int) *Sink {
	if maxRetries < 1 {
		maxRetries = 1
	}
	return &Sink{store: store, dedup: NewDedup(), maxRetries: maxRetries}
}

// WriteResult 写入窗口差异；已提交位点之前与已落库的差异自动跳过。
// 单条差异写入失败时会按 maxRetries 重试；若最终仍失败，则真实上报该错误（而非吞掉），
// 以便上层据此重试整个窗口，避免差异静默丢失。
func (s *Sink) WriteResult(ctx context.Context, window model.Window, result *model.Result, resume ResumeInfo) error {
	for _, entry := range result.Entries {
		key := entry.DedupKey()
		if s.dedup.Seen(key) {
			continue
		}
		if err := s.putWithRetry(ctx, entry); err != nil {
			return fmt.Errorf("write diff %s for %s: %w", key, window.ID, err)
		}
		s.dedup.Mark(key)
	}
	return nil
}

// putWithRetry 写入单条差异，失败时按 maxRetries 重试。
// 重试期间若上下文已取消则立即返回，避免无意义的重试。
func (s *Sink) putWithRetry(ctx context.Context, entry model.Entry) error {
	var lastErr error
	for attempt := 0; attempt <= s.maxRetries; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := s.store.Put(ctx, entry); err != nil {
			lastErr = err
			if attempt < s.maxRetries {
				select {
				case <-time.After(retryBackoff):
				case <-ctx.Done():
					return ctx.Err()
				}
				continue
			}
			return lastErr
		}
		return nil
	}
	return lastErr
}
