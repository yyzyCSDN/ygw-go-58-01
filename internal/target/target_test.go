package target

import (
	"context"
	"testing"

	"reconcilesvc/internal/model"
)

func TestMemoryStoreSnapshot(t *testing.T) {
	store := NewMemoryStore(7, map[string]*model.Record{"k": {Key: "k", Version: 2}})
	if store.Version() != 7 {
		t.Fatalf("version = %d", store.Version())
	}
	snap, err := store.Snapshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if snap.Version != 7 || len(snap.Records) != 1 {
		t.Fatalf("snapshot mismatch: %+v", snap)
	}
}

func TestBaselineGet(t *testing.T) {
	store := NewMemoryStore(1, map[string]*model.Record{"a": {Key: "a", Version: 1}})
	manager := NewBaselineManager(store)
	base, err := manager.Acquire(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if rec, ok := base.Get("a"); !ok || rec.Version != 1 {
		t.Fatalf("baseline Get(a) = %v, %v", rec, ok)
	}
	if _, ok := base.Get("nope"); ok {
		t.Fatalf("baseline Get(nope) should be missing")
	}
}
