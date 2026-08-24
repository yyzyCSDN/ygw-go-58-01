package diff

import (
	"context"

	"reconcilesvc/internal/model"
)

// Filter 描述差异查询过滤条件。
type Filter struct {
	Kind  string
	Limit int
}

// ListFiltered 按类型过滤差异列表。
func ListFiltered(ctx context.Context, store Store, filter Filter) ([]model.Entry, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 200
	}
	entries, err := store.List(ctx, limit)
	if err != nil {
		return nil, err
	}
	if filter.Kind == "" {
		return entries, nil
	}
	out := make([]model.Entry, 0, len(entries))
	for _, entry := range entries {
		if entry.Kind.String() == filter.Kind {
			out = append(out, entry)
		}
	}
	return out, nil
}
