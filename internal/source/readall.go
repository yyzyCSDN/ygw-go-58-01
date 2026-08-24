package source

import (
	"context"
	"fmt"
	"io"

	"reconcilesvc/internal/model"
)

// ReadWindow 分块读取数据源，返回窗口区间内的记录块。
func ReadWindow(ctx context.Context, s Source, w model.Window, chunkSize int) ([][]*model.Record, error) {
	chunks, err := readWindowInternal(ctx, s, w, chunkSize, nil)
	if err != nil {
		return nil, fmt.Errorf("read window %s: %w", w.ID, err)
	}
	return chunks, nil
}

// ReadWindowTracked 分块读取数据源并记录读取统计。
func ReadWindowTracked(ctx context.Context, s Source, w model.Window, chunkSize int, tracker *ReadTracker) ([][]*model.Record, error) {
	return readWindowInternal(ctx, s, w, chunkSize, tracker)
}

func readWindowInternal(ctx context.Context, s Source, w model.Window, chunkSize int, tracker *ReadTracker) ([][]*model.Record, error) {
	reader, err := s.Open(ctx)
	if err != nil {
		return nil, err
	}
	// 读取结束（成功或出错）都必须关闭读取器以归还连接句柄，
	// 否则源端连接数持续累积，最终触发连接告警并导致读取失败。
	defer func() { _ = reader.Close() }()
	chunker := NewChunker(reader, chunkSize, s.Total())
	var chunks [][]*model.Record
	stat := ReadStat{WindowID: w.ID}
	for {
		chunk, err := chunker.NextChunk(ctx)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		stat.Read += len(chunk)
		stat.Chunks++
		kept := make([]*model.Record, 0, len(chunk))
		for _, rec := range chunk {
			if w.Contains(rec.Key) {
				kept = append(kept, rec)
			}
		}
		stat.Kept += len(kept)
		if len(kept) > 0 {
			chunks = append(chunks, kept)
		}
	}
	if !chunker.Complete() {
		return nil, fmt.Errorf("window %s read incomplete: %d/%d records", w.ID, chunker.Cursor(), s.Total())
	}
	for _, chunk := range chunks {
		for _, rec := range chunk {
			if !w.Contains(rec.Key) {
				return nil, fmt.Errorf("record %s outside window %s", rec.Key, w.ID)
			}
		}
	}
	if tracker != nil {
		stat.Partial = chunker.Partial()
		tracker.Record(stat)
	}
	return chunks, nil
}
