package source

import (
	"context"
	"io"

	"reconcilesvc/internal/model"
)

// Chunker 从读取器按固定大小切分记录块，边界记录完整保留。
type Chunker struct {
	reader Reader
	size   int
	cursor int
	total  int
	done   bool
}

// NewChunker 创建分块器。
func NewChunker(reader Reader, size, total int) *Chunker {
	return &Chunker{reader: reader, size: size, total: total}
}

// NextChunk 返回下一块记录；全部读完返回 io.EOF。
func (c *Chunker) NextChunk(ctx context.Context) ([]*model.Record, error) {
	if c.done {
		return nil, io.EOF
	}
	chunk := make([]*model.Record, 0, c.size)
	for len(chunk) < c.size {
		rec, err := c.reader.Next(ctx)
		if err == io.EOF {
			// 读取器已耗尽：先冲刷已累积的尾块，下一次调用再返回 EOF，
			// 确保块边界上的记录不会随 EOF 被丢弃。
			c.done = true
			if len(chunk) > 0 {
				return chunk, nil
			}
			return nil, io.EOF
		}
		if err != nil {
			return nil, err
		}
		chunk = append(chunk, rec)
		c.cursor++
	}
	return chunk, nil
}

// Cursor 返回已消费的记录数。
func (c *Chunker) Cursor() int { return c.cursor }

// Complete 表示读取器已被完整消费。
func (c *Chunker) Complete() bool { return c.done && c.cursor == c.total }

// Partial 表示最后一个块是否是未满的尾块。
func (c *Chunker) Partial() bool { return c.done && c.cursor < c.total }

// Close 关闭底层读取器。
func (c *Chunker) Close() error { return c.reader.Close() }
