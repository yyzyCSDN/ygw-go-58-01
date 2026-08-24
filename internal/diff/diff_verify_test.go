package diff

import (
	"context"
	"errors"
	"testing"

	"reconcilesvc/internal/model"
)

type alwaysFailingStore struct{}

func (alwaysFailingStore) Put(ctx context.Context, entry model.Entry) error {
	return errors.New("disk full")
}

func (alwaysFailingStore) Has(ctx context.Context, key string) (bool, error) {
	return false, nil
}

func (alwaysFailingStore) List(ctx context.Context, limit int) ([]model.Entry, error) {
	return nil, nil
}

func TestDiffWriteErrorNotSwallowed(t *testing.T) {
	sink := NewSink(alwaysFailingStore{}, 1)
	result := &model.Result{
		Entries: []model.Entry{{Key: "k1", Kind: model.DiffMismatch}},
	}
	err := sink.WriteResult(context.Background(), model.Window{ID: "w"}, result, ResumeInfo{})
	if err == nil {
		t.Fatalf("diff write failure was swallowed, diffs are lost")
	}
}
