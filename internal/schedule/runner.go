package schedule

import (
	"context"
	"sync"
)

// Runner 用固定数量 worker 并发执行对账任务。
type Runner struct {
	workers int
}

// NewRunner 创建任务执行器。
func NewRunner(workers int) *Runner {
	if workers <= 0 {
		workers = 1
	}
	return &Runner{workers: workers}
}

// Run 并发执行 count 个任务函数，任一失败即返回第一个错误。
func (r *Runner) Run(ctx context.Context, count int, fn func(ctx context.Context, index int) error) error {
	sem := make(chan struct{}, r.workers)
	errs := make([]error, count)
	var wg sync.WaitGroup
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				errs[index] = ctx.Err()
				return
			}
			defer func() { <-sem }()
			errs[index] = fn(ctx, index)
		}(i)
	}
	wg.Wait()
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}
