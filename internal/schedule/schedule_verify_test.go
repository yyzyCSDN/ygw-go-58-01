package schedule

import (
	"context"
	"fmt"
	"testing"
	"time"

	"reconcilesvc/internal/compare"
	"reconcilesvc/internal/diff"
	"reconcilesvc/internal/model"
	"reconcilesvc/internal/notify"
	"reconcilesvc/internal/offset"
	"reconcilesvc/internal/source"
	"reconcilesvc/internal/target"
)

func TestReconcileWindowsNoGap(t *testing.T) {
	pool := source.NewConnPool()
	records := make([]*model.Record, 0, 130)
	keys := make([]string, 0, 130)
	for i := 0; i < 130; i++ {
		key := fmt.Sprintf("key-%03d", i)
		keys = append(keys, key)
		records = append(records, &model.Record{Key: key, Version: int64(i % 7), Data: "d"})
	}
	targetStore := target.NewMemoryStore(1, map[string]*model.Record{})
	baseline := target.NewBaselineManager(targetStore)
	diffStore := diff.NewMemoryStore()
	sink := diff.NewSink(diffStore, 1)
	engine := compare.NewEngine(baseline, sink, compare.DefaultOptions())
	offsets := offset.NewManager(offset.NewMemoryStore())
	counting := notify.NewCountingSubscriber()
	notifier := notify.NewNotifier(counting)
	planner := NewPlanner(50)
	scheduler := NewScheduler(
		planner,
		source.NewMemorySource("src", records, pool),
		engine,
		offsets,
		notifier,
		16,
		5*time.Second,
		NewRetryPolicy(2),
		keys,
		source.NewReadTracker(),
	)
	task := &Task{ID: "t1"}
	result, err := scheduler.Run(context.Background(), task)
	if err != nil {
		t.Fatalf("reconcile failed: %v", err)
	}
	if task.State != StateDone {
		t.Fatalf("task state = %s, want done", task.State)
	}
	if result.Missing != 130 {
		t.Fatalf("reconciled %d keys, want 130 (window gap dropped a window)", result.Missing)
	}
	if pos := offsets.Position(); pos.Key != "key-129" {
		t.Fatalf("offset did not reach the final key: %+v", pos)
	}
}
