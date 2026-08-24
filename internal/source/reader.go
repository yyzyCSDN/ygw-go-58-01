package source

import (
	"context"
	"fmt"
	"io"

	"reconcilesvc/internal/model"
)

// sliceReader 按切片顺序输出记录，并在 Close 时归还连接。
type sliceReader struct {
	records []*model.Record
	index   int
	conn    *Conn
	closed  bool
}

func newSliceReader(records []*model.Record, conn *Conn) *sliceReader {
	return &sliceReader{records: records, conn: conn}
}

// Next 返回下一条记录；ctx 取消时立即返回错误。
func (r *sliceReader) Next(ctx context.Context) (*model.Record, error) {
	if r.closed {
		return nil, fmt.Errorf("read from closed reader")
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	if r.index >= len(r.records) {
		return nil, io.EOF
	}
	rec := r.records[r.index]
	r.index++
	return rec, nil
}

// Closed 返回读取器是否已关闭。
func (r *sliceReader) Closed() bool { return r.closed }

// Close 幂等关闭读取器并归还连接。
func (r *sliceReader) Close() error {
	r.closed = true
	return nil
}
