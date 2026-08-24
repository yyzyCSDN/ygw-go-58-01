package diff

import (
	"context"
	"errors"
	"testing"

	"reconcilesvc/internal/model"
)

func TestMemoryStoreDedup(t *testing.T) {
	store := NewMemoryStore()
	entry := model.Entry{Key: "k", Kind: model.DiffMissing}
	if err := store.Put(context.Background(), entry); err != nil {
		t.Fatal(err)
	}
	if err := store.Put(context.Background(), entry); err != nil {
		t.Fatal(err)
	}
	entries, err := store.List(context.Background(), 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("stored %d entries, want 1", len(entries))
	}
	ok, err := store.Has(context.Background(), entry.DedupKey())
	if err != nil || !ok {
		t.Fatalf("Has(%q) = %v, %v", entry.DedupKey(), ok, err)
	}
}

func TestMemoryStoreFailNext(t *testing.T) {
	store := NewMemoryStore()
	store.FailNext(errors.New("disk full"))
	if err := store.Put(context.Background(), model.Entry{Key: "k"}); err == nil {
		t.Fatalf("expected injected failure")
	}
}
