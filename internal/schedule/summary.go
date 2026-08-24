package schedule

import (
	"context"
	"errors"

	"reconcilesvc/internal/compare"
	"reconcilesvc/internal/model"
)

// BatchResult 汇总一次批量对账。
type BatchResult struct {
	Tasks int
	Total *model.Result
}

// RunBatch 用 Runner 并发执行多个任务并聚合结果。
func RunBatch(ctx context.Context, runner *Runner, scheduler *Scheduler, tasks []*Task) (*BatchResult, error) {
	if len(tasks) == 0 {
		return &BatchResult{}, nil
	}
	results := make([]*model.Result, len(tasks))
	err := runner.Run(ctx, len(tasks), func(rctx context.Context, index int) error {
		result, err := scheduler.Run(rctx, tasks[index])
		if err != nil {
			return err
		}
		results[index] = result
		return nil
	})
	total := &model.Result{}
	for _, result := range results {
		compare.Merge(total, result)
	}
	batch := &BatchResult{Tasks: len(tasks), Total: total}
	if err != nil {
		return batch, errors.Join(err)
	}
	return batch, nil
}
