package source

import (
	"context"
	"io"
)

// Probe 读取源的首条记录以验证源可用性。
func Probe(ctx context.Context, s Source) error {
	reader, err := s.Open(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = reader.Close()
	}()
	_, err = reader.Next(ctx)
	if err != nil && err != io.EOF {
		return err
	}
	return nil
}
