package compare

import (
	"context"
	"testing"

	"reconcilesvc/internal/diff"
	"reconcilesvc/internal/model"
	"reconcilesvc/internal/target"
)

// TestReconcileReflectsTargetUpdate 验证目标数据更新后，复用同一引擎再次对账能检出
// 最新基线带来的差异，而不是沿用旧版本基线把记录判为一致。
func TestReconcileReflectsTargetUpdate(t *testing.T) {
	store := target.NewMemoryStore(1, map[string]*model.Record{
		"k1": {Key: "k1", Version: 1, Data: "v1"},
	})
	engine := NewEngine(target.NewBaselineManager(store), &fakeSink{}, DefaultOptions())
	records := []*model.Record{{Key: "k1", Version: 1, Data: "v1"}}

	// 初始：源与目标一致，无差异。
	r1, err := engine.ReconcileWindow(context.Background(), model.Window{ID: "w"}, records, diff.ResumeInfo{})
	if err != nil {
		t.Fatal(err)
	}
	if r1.Mismatch != 0 {
		t.Fatalf("before update: mismatch = %d, want 0", r1.Mismatch)
	}

	// 目标更新：k1 数据变化，并新增 k2（源缺失）。
	store.Update(2, map[string]*model.Record{
		"k1": {Key: "k1", Version: 2, Data: "v2"},
		"k2": {Key: "k2", Version: 1, Data: "new"},
	})

	// 复用同一引擎再次对账，应基于最新基线检出 k1 不一致与 k2 多余。
	r2, err := engine.ReconcileWindow(context.Background(), model.Window{ID: "w"}, records, diff.ResumeInfo{})
	if err != nil {
		t.Fatal(err)
	}
	if r2.Mismatch != 1 {
		t.Fatalf("after update: mismatch = %d, want 1 (target update not detected)", r2.Mismatch)
	}
	if r2.Extra != 1 {
		t.Fatalf("after update: extra = %d, want 1 (new target key not detected)", r2.Extra)
	}
}

// TestAcquireRefreshesOnVersionChange 直接验证 BaselineManager 在目标版本变化后
// 自动刷新基线，而非返回旧快照。
func TestAcquireRefreshesOnVersionChange(t *testing.T) {
	store := target.NewMemoryStore(1, map[string]*model.Record{
		"k1": {Key: "k1", Version: 1, Data: "v1"},
	})
	manager := target.NewBaselineManager(store)

	base1, err := manager.Acquire(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if base1.Version() != 1 {
		t.Fatalf("base1 version = %d, want 1", base1.Version())
	}

	// 目标更新到版本 2。
	store.Update(2, map[string]*model.Record{
		"k1": {Key: "k1", Version: 2, Data: "v2"},
	})

	base2, err := manager.Acquire(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if base2.Version() != 2 {
		t.Fatalf("base2 version = %d, want 2 (stale baseline returned)", base2.Version())
	}
	if rec, ok := base2.Get("k1"); !ok || rec.Version != 2 || rec.Data != "v2" {
		t.Fatalf("base2 Get(k1) = %v, %v, want updated record", rec, ok)
	}
}
