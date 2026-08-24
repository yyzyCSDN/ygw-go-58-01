package schedule

import (
	"context"
	"errors"
	"fmt"
	"time"

	"reconcilesvc/internal/compare"
	"reconcilesvc/internal/model"
	"reconcilesvc/internal/notify"
	"reconcilesvc/internal/offset"
	"reconcilesvc/internal/source"
)

// Scheduler 编排对账任务的全流程。
type Scheduler struct {
	planner   *Planner
	source    source.Source
	engine    *compare.Engine
	offsets   *offset.Manager
	notifier  *notify.Notifier
	chunkSize int
	timeout   time.Duration
	retry     RetryPolicy
	keys      []string
	tracker   *source.ReadTracker
}

// NewScheduler 创建调度器。
func NewScheduler(
	planner *Planner,
	src source.Source,
	engine *compare.Engine,
	offsets *offset.Manager,
	notifier *notify.Notifier,
	chunkSize int,
	timeout time.Duration,
	retry RetryPolicy,
	keys []string,
	tracker *source.ReadTracker,
) *Scheduler {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &Scheduler{
		planner:   planner,
		source:    src,
		engine:    engine,
		offsets:   offsets,
		notifier:  notifier,
		chunkSize: chunkSize,
		timeout:   timeout,
		retry:     retry,
		keys:      keys,
		tracker:   tracker,
	}
}

// Plan 返回从当前位点续排的窗口。
//
// 只要位点已推进到某个键，就只返回该键之后尚未对账的窗口，避免下一周期
// 重复对账已完成的窗口；位点为空时才从零开始规划全量。
func (s *Scheduler) Plan(phase model.Phase) []model.Window {
	pos := s.offsets.Position()
	if pos.Key == "" {
		return s.planner.Plan(s.keys, phase)
	}
	return s.planner.ResumeWindows(s.keys, phase, pos.Key)
}

// FullWindows 返回从零开始的全量窗口，用于切换增量的锚点。
func (s *Scheduler) FullWindows() []model.Window {
	return s.planner.Plan(s.keys, model.PhaseFull)
}

// Run 执行任务：逐窗口读取、比对、写差异、推进位点，最后通知并完成。
func (s *Scheduler) Run(ctx context.Context, task *Task) (*model.Result, error) {
	task.State = StateRunning
	taskCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	if task.Windows == nil {
		task.Windows = s.Plan(model.PhaseFull)
	}
	total := &model.Result{Window: model.Window{ID: task.ID}, StartedAt: time.Now()}
	for i, window := range task.Windows {
		if err := taskCtx.Err(); err != nil {
			task.State = StatePartial
			task.Err = err
			return total, err
		}
		task.Current = i
		windowCtx, windowCancel := s.windowContext(taskCtx)
		result, err := s.runWindow(windowCtx, window)
		windowCancel()
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				err = fmt.Errorf("window %s timed out after %s: %w", window.ID, s.timeout, err)
			}
			if s.retry.ShouldRetry(task.Retries) {
				task.Retries++
				task.State = StatePartial
				return total, err
			}
			task.State = StatePartial
			task.Err = err
			return total, err
		}
		compare.Merge(total, result)
		if err := s.offsets.Advance(taskCtx, window); err != nil {
			task.State = StatePartial
			task.Err = err
			return total, err
		}
		if err := s.notifier.Notify(taskCtx, window, result, s.offsets.Position().Key); err != nil {
			task.State = StatePartial
			task.Err = err
			return total, err
		}
	}
	if len(task.Windows) > 0 {
		if err := offset.Continuity(task.Windows, s.keys); err != nil {
			task.State = StatePartial
			task.Err = err
			return total, err
		}
		if err := s.offsets.Complete(taskCtx, task.Windows[len(task.Windows)-1]); err != nil {
			task.State = StatePartial
			task.Err = err
			return total, err
		}
	}
	task.State = StateDone
	total.FinishedAt = time.Now()
	return total, nil
}

// windowContext 创建窗口级超时上下文。
func (s *Scheduler) windowContext(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, s.timeout)
}

// runWindow 读取一个窗口的源数据并比对。
func (s *Scheduler) runWindow(ctx context.Context, window model.Window) (*model.Result, error) {
	chunks, err := source.ReadWindowTracked(ctx, s.source, window, s.chunkSize, s.tracker)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", window.ID, err)
	}
	records := flatten(chunks)
	resume, err := s.offsets.ResumeInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("resume info: %w", err)
	}
	return s.engine.ReconcileWindow(ctx, window, records, resume)
}

// flatten 将分块记录展开为单层记录。
func flatten(chunks [][]*model.Record) []*model.Record {
	var out []*model.Record
	for _, chunk := range chunks {
		out = append(out, chunk...)
	}
	return out
}
