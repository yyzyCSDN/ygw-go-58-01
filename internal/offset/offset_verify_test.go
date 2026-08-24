package offset

import (
	"context"
	"testing"

	"reconcilesvc/internal/model"
)

func TestOffsetAdvancedAfterReconcile(t *testing.T) {
	store := NewMemoryStore()
	manager := NewManager(store)
	w0 := model.Window{ID: "w0", Index: 0, End: "key-049"}
	w1 := model.Window{ID: "w1", Index: 1, End: "key-099"}
	if err := manager.Advance(context.Background(), w0); err != nil {
		t.Fatal(err)
	}
	if err := manager.Advance(context.Background(), w1); err != nil {
		t.Fatal(err)
	}
	pos := manager.Position()
	if pos.WindowIndex != 1 || pos.Key != "key-099" {
		t.Fatalf("offset was not advanced after reconcile: %+v", pos)
	}
	stored, err := store.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if stored.Key != "key-099" || stored.WindowIndex != 1 {
		t.Fatalf("persisted offset was not advanced: %+v", stored)
	}
	fresh := NewManager(store)
	info, err := fresh.ResumeInfo(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if info.CommittedKey != "key-099" {
		t.Fatalf("restart resumed from %q, want key-099", info.CommittedKey)
	}
}
