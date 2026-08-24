package source

import (
	"context"
	"io"
	"testing"

	"reconcilesvc/internal/model"
)

func records(n int) []*model.Record {
	out := make([]*model.Record, 0, n)
	for i := 0; i < n; i++ {
		key := string(rune('a' + i))
		out = append(out, &model.Record{Key: key, Version: int64(i), Data: "d"})
	}
	return out
}

func TestChunkerExactBlocks(t *testing.T) {
	pool := NewConnPool()
	reader, err := NewMemorySource("s", records(4), pool).Open(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = reader.Close() }()
	chunker := NewChunker(reader, 2, 4)
	chunk, err := chunker.NextChunk(context.Background())
	if err != nil || len(chunk) != 2 {
		t.Fatalf("first chunk = %d, %v", len(chunk), err)
	}
	chunk, err = chunker.NextChunk(context.Background())
	if err != nil || len(chunk) != 2 {
		t.Fatalf("second chunk = %d, %v", len(chunk), err)
	}
	if _, err := chunker.NextChunk(context.Background()); err != io.EOF {
		t.Fatalf("expected EOF, got %v", err)
	}
	if !chunker.Complete() {
		t.Fatalf("chunker should be complete")
	}
}

func TestBucketRange(t *testing.T) {
	for _, key := range []string{"a", "b", "reconcile-key", "zzz"} {
		if b := Bucket(key, 8); b < 0 || b >= 8 {
			t.Fatalf("bucket %q out of range: %d", key, b)
		}
	}
}

func TestConnPoolRelease(t *testing.T) {
	pool := NewConnPool()
	conn, err := pool.Acquire(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if pool.Active() != 1 {
		t.Fatalf("active = %d, want 1", pool.Active())
	}
	if err := conn.release(); err != nil {
		t.Fatal(err)
	}
	if pool.Active() != 0 {
		t.Fatalf("active = %d, want 0", pool.Active())
	}
	if err := conn.release(); err != nil {
		t.Fatal(err)
	}
}
