package offset

import (
	"context"
	"testing"

	"reconcilesvc/internal/model"
)

func TestMemoryStoreRoundTrip(t *testing.T) {
	store := NewMemoryStore()
	pos := model.Position{Phase: model.PhaseIncremental, WindowIndex: 3, Key: "k", Completed: true}
	if err := store.Save(context.Background(), pos); err != nil {
		t.Fatal(err)
	}
	got, err := store.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got != pos {
		t.Fatalf("loaded %+v, want %+v", got, pos)
	}
}

func TestManagerPositionZero(t *testing.T) {
	manager := NewManager(NewMemoryStore())
	if pos := manager.Position(); pos.Completed || pos.Key != "" {
		t.Fatalf("initial position = %+v", pos)
	}
}

func TestAnchorWindow(t *testing.T) {
	windows := []model.Window{
		{ID: "w0", Index: 0, End: "m"},
		{ID: "w1", Index: 1, End: "z"},
	}
	anchor, err := AnchorWindow(windows)
	if err != nil {
		t.Fatal(err)
	}
	if anchor.Index != 1 || anchor.End != "z" {
		t.Fatalf("anchor = %+v", anchor)
	}
}
