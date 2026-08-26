package compare

import (
	"context"
	"testing"

	"reconcilesvc/internal/diff"
	"reconcilesvc/internal/model"
	"reconcilesvc/internal/target"
)

func TestMissingKeyNoNilPanic(t *testing.T) {
	store := target.NewMemoryStore(1, map[string]*model.Record{
		"present": {Key: "present", Version: 1, Data: "p"},
		"broken":  nil,
	})
	manager := target.NewBaselineManager(store)
	engine := NewEngine(manager, &fakeSink{}, DefaultOptions())
	records := []*model.Record{
		{Key: "present", Version: 1, Data: "p"},
		{Key: "broken", Version: 2, Data: "b"},
	}
	defer func() {
		if recover() != nil {
			t.Fatalf("panicked while reconciling a missing key")
		}
	}()
	result, err := engine.ReconcileWindow(context.Background(), model.Window{ID: "w"}, records, diff.ResumeInfo{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Missing != 1 {
		t.Fatalf("missing = %d, want 1", result.Missing)
	}
}
