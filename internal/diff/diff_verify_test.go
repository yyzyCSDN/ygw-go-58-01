package diff_test

import (
	"context"
	"testing"

	"reconcilesvc/internal/diff"
	"reconcilesvc/internal/model"
	"reconcilesvc/internal/offset"
)

func TestReconcileRetryNoDuplicateDiff(t *testing.T) {
	offsetStore := offset.NewMemoryStore()
	offsets := offset.NewManager(offsetStore)
	w0 := model.Window{ID: "w0", Index: 0, End: "key-001"}
	if err := offsets.Advance(context.Background(), w0); err != nil {
		t.Fatal(err)
	}
	resume, err := offsets.ResumeInfo(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	store := diff.NewMemoryStore()
	first := diff.NewSink(store, 1)
	entries := []model.Entry{
		{Key: "key-000", Kind: model.DiffMismatch},
		{Key: "key-001", Kind: model.DiffMismatch},
		{Key: "key-002", Kind: model.DiffMismatch},
	}
	result := &model.Result{Entries: entries}
	if err := first.WriteResult(context.Background(), model.Window{ID: "w"}, result, resume); err != nil {
		t.Fatal(err)
	}

	// 模拟重启后的重试：全新 Sink 实例再次写入同一结果。
	second := diff.NewSink(store, 1)
	if err := second.WriteResult(context.Background(), model.Window{ID: "w"}, result, resume); err != nil {
		t.Fatal(err)
	}

	got, err := store.List(context.Background(), 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("retry duplicated diffs: %d entries, want 1", len(got))
	}
}
