package compare

import (
	"context"
	"fmt"
	"sync"
	"time"

	"reconcilesvc/internal/diff"
	"reconcilesvc/internal/model"
	"reconcilesvc/internal/target"
)

// BaselineProvider 提供最新目标基线。
type BaselineProvider interface {
	Acquire(ctx context.Context) (*target.Baseline, error)
}

// DiffSink 接收对账结果。
type DiffSink interface {
	WriteResult(ctx context.Context, window model.Window, result *model.Result, resume diff.ResumeInfo) error
}

// Engine 执行逐键比对。
type Engine struct {
	baseline    BaselineProvider
	sink        DiffSink
	opts        Options
	mu          sync.Mutex
	seenVersion int64
}

// NewEngine 创建比对引擎。
func NewEngine(baseline BaselineProvider, sink DiffSink, opts Options) *Engine {
	return &Engine{baseline: baseline, sink: sink, opts: opts}
}

// ReconcileWindow 将源记录与最新目标基线比对，并写入差异。
func (e *Engine) ReconcileWindow(ctx context.Context, window model.Window, records []*model.Record, resume diff.ResumeInfo) (*model.Result, error) {
	base, err := e.baseline.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("acquire baseline for %s: %w", window.ID, err)
	}
	e.mu.Lock()
	if base.Version() < e.seenVersion {
		e.mu.Unlock()
		return nil, fmt.Errorf("baseline version regressed from %d to %d", e.seenVersion, base.Version())
	}
	e.seenVersion = base.Version()
	e.mu.Unlock()
	result := &model.Result{Window: window, StartedAt: time.Now()}
	seen := make(map[string]struct{}, len(records))
	for _, rec := range records {
		if rec == nil {
			return nil, fmt.Errorf("%s contains nil record", window.ID)
		}
		result.Total++
		seen[rec.Key] = struct{}{}
		targetRec, ok := base.Get(rec.Key)
		if !ok || targetRec == nil {
			result.Entries = append(result.Entries, missingEntry(rec))
			result.Missing++
			continue
		}
		versionSame := rec.Version == targetRec.Version
		dataSame := !e.opts.CompareData || rec.Data == targetRec.Data
		if !versionSame || !dataSame {
			result.Entries = append(result.Entries, model.Entry{
				Key:           rec.Key,
				Kind:          model.DiffMismatch,
				SourceVersion: rec.Version,
				TargetVersion: targetRec.Version,
				SourceData:    rec.Data,
				TargetData:    targetRec.Data,
			})
			result.Mismatch++
		}
	}
	for _, key := range base.Keys() {
		if _, ok := seen[key]; ok {
			continue
		}
		rec, _ := base.Get(key)
		entry := model.Entry{Key: key, Kind: model.DiffExtra}
		if rec != nil {
			entry.TargetVersion = rec.Version
			entry.TargetData = rec.Data
		}
		result.Entries = append(result.Entries, entry)
		result.Extra++
	}
	if err := e.sink.WriteResult(ctx, window, result, resume); err != nil {
		return nil, fmt.Errorf("write result for %s: %w", window.ID, err)
	}
	result.FinishedAt = time.Now()
	return result, nil
}

// missingEntry 构造一条缺失差异条目；空记录按空差异处理。
func missingEntry(rec *model.Record) model.Entry {
	if rec == nil {
		return model.Entry{}
	}
	return model.Entry{
		Key:           rec.Key,
		Kind:          model.DiffMissing,
		SourceVersion: rec.Version,
		SourceData:    rec.Data,
	}
}
