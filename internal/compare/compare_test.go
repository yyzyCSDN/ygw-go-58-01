package compare

import (
	"context"
	"testing"

	"reconcilesvc/internal/diff"
	"reconcilesvc/internal/model"
	"reconcilesvc/internal/target"
)

type fakeSink struct {
	written []*model.Result
}

func (f *fakeSink) WriteResult(ctx context.Context, window model.Window, result *model.Result, resume diff.ResumeInfo) error {
	f.written = append(f.written, result)
	return nil
}

func TestReconcileDetectsMismatch(t *testing.T) {
	store := target.NewMemoryStore(1, map[string]*model.Record{
		"k1": {Key: "k1", Version: 2, Data: "new"},
	})
	engine := NewEngine(target.NewBaselineManager(store), &fakeSink{}, DefaultOptions())
	records := []*model.Record{{Key: "k1", Version: 2, Data: "old"}}
	result, err := engine.ReconcileWindow(context.Background(), model.Window{ID: "w"}, records, diff.ResumeInfo{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Mismatch != 1 {
		t.Fatalf("mismatch = %d, want 1", result.Mismatch)
	}
}

func TestReconcileDetectsMissing(t *testing.T) {
	store := target.NewMemoryStore(1, map[string]*model.Record{})
	engine := NewEngine(target.NewBaselineManager(store), &fakeSink{}, DefaultOptions())
	records := []*model.Record{{Key: "only-source", Version: 1}}
	result, err := engine.ReconcileWindow(context.Background(), model.Window{ID: "w"}, records, diff.ResumeInfo{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Missing != 1 {
		t.Fatalf("missing = %d, want 1", result.Missing)
	}
}
