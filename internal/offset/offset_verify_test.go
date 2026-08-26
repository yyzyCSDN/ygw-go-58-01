package offset_test

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
	"reconcilesvc/internal/schedule"
	"reconcilesvc/internal/source"
	"reconcilesvc/internal/target"
)

func TestFullIncrementalHandoffOffset(t *testing.T) {
	pool := source.NewConnPool()
	records := make([]*model.Record, 0, 200)
	keys := make([]string, 0, 200)
	for i := 0; i < 200; i++ {
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
	planner := schedule.NewPlanner(50)
	scheduler := schedule.NewScheduler(
		planner,
		source.NewMemorySource("src", records, pool),
		engine,
		offsets,
		notifier,
		16,
		5*time.Second,
		schedule.NewRetryPolicy(2),
		keys,
		source.NewReadTracker(),
	)

	full := planner.Plan(keys, model.PhaseFull)
	anchor := full[1]
	if err := offsets.SwitchToIncremental(context.Background(), anchor); err != nil {
		t.Fatal(err)
	}
	pos := offsets.Position()
	if pos.Phase != model.PhaseIncremental || pos.WindowIndex != anchor.Index || pos.Key != anchor.End {
		t.Fatalf("handoff position misaligned: %+v (anchor %+v)", pos, anchor)
	}
	incremental := scheduler.Plan(model.PhaseIncremental)
	if len(incremental) == 0 {
		t.Fatalf("incremental windows are empty after handoff")
	}
	if incremental[0].Start <= anchor.End {
		t.Fatalf("incremental overlaps the handoff window: start=%s anchor=%s", incremental[0].Start, anchor.End)
	}
}
