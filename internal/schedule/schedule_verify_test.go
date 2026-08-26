package schedule

import (
	"context"
	"io"
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

type blockingReader struct{ stop chan struct{} }

func (r *blockingReader) Next(ctx context.Context) (*model.Record, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-r.stop:
		return nil, io.EOF
	}
}

func (r *blockingReader) Close() error { return nil }
func (r *blockingReader) Closed() bool { return false }

type slowSource struct{ reader *blockingReader }

func (s *slowSource) Open(ctx context.Context) (source.Reader, error) { return s.reader, nil }
func (s *slowSource) Total() int                                     { return 100 }
func (s *slowSource) Name() string                                   { return "slow" }

func TestReconcileHonorsTimeoutContext(t *testing.T) {
	keys := []string{"key-000", "key-001", "key-002", "key-003"}
	targetStore := target.NewMemoryStore(1, map[string]*model.Record{})
	baseline := target.NewBaselineManager(targetStore)
	diffStore := diff.NewMemoryStore()
	sink := diff.NewSink(diffStore, 1)
	engine := compare.NewEngine(baseline, sink, compare.DefaultOptions())
	offsets := offset.NewManager(offset.NewMemoryStore())
	counting := notify.NewCountingSubscriber()
	notifier := notify.NewNotifier(counting)
	slow := &slowSource{reader: &blockingReader{stop: make(chan struct{})}}
	scheduler := NewScheduler(
		NewPlanner(2),
		slow,
		engine,
		offsets,
		notifier,
		16,
		100*time.Millisecond,
		NewRetryPolicy(1),
		keys,
		nil,
	)
	task := &Task{ID: "slow"}
	done := make(chan struct{})
	var runErr error
	go func() {
		_, runErr = scheduler.Run(context.Background(), task)
		close(done)
	}()
	select {
	case <-done:
		if runErr == nil {
			t.Fatalf("expected timeout error for the slow source")
		}
		if task.State != StatePartial {
			t.Fatalf("task state = %s, want partial", task.State)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("slow source hung the worker; timeout context was not honored")
	}
}
