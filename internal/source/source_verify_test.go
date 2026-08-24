package source

import (
	"context"
	"fmt"
	"testing"

	"reconcilesvc/internal/model"
)

func TestSourceReaderClosed(t *testing.T) {
	pool := NewConnPool()
	records := make([]*model.Record, 0, 130)
	for i := 0; i < 130; i++ {
		key := fmt.Sprintf("key-%03d", i)
		records = append(records, &model.Record{Key: key, Version: int64(i), Data: "d"})
	}
	src := NewMemorySource("src", records, pool)
	window := model.Window{ID: "all", Start: "key-000", End: "key-129"}
	for i := 0; i < 3; i++ {
		chunks, err := ReadWindow(context.Background(), src, window, 16)
		if err != nil {
			t.Fatal(err)
		}
		total := 0
		for _, chunk := range chunks {
			total += len(chunk)
		}
		if total != 130 {
			t.Fatalf("read %d records, want 130", total)
		}
		if pool.Active() != 0 {
			t.Fatalf("source reader leaked connection: active=%d", pool.Active())
		}
	}
}
