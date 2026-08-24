package compare

import (
	"context"
	"testing"

	"reconcilesvc/internal/diff"
	"reconcilesvc/internal/model"
	"reconcilesvc/internal/target"
)

func TestCompareUsesLatestBaseline(t *testing.T) {
	store := target.NewMemoryStore(1, map[string]*model.Record{
		"k1": {Key: "k1", Version: 1, Data: "a"},
	})
	manager := target.NewBaselineManager(store)
	engine := NewEngine(manager, &fakeSink{}, DefaultOptions())
	records := []*model.Record{{Key: "k1", Version: 1, Data: "a"}}
	if _, err := engine.ReconcileWindow(context.Background(), model.Window{ID: "w1"}, records, diff.ResumeInfo{}); err != nil {
		t.Fatal(err)
	}
	store.Update(2, map[string]*model.Record{
		"k1": {Key: "k1", Version: 2, Data: "b"},
	})
	result, err := engine.ReconcileWindow(context.Background(), model.Window{ID: "w2"}, records, diff.ResumeInfo{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Mismatch != 1 {
		t.Fatalf("mismatch = %d, want 1 (stale baseline hid the update)", result.Mismatch)
	}
}
