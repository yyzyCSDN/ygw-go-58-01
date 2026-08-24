package diff

import "sync"

// Dedup 记录本次写入的差异键。
type Dedup struct {
	mu   sync.Mutex
	seen map[string]struct{}
}

// NewDedup 创建去重器。
func NewDedup() *Dedup {
	return &Dedup{seen: make(map[string]struct{})}
}

// Seen 判断键是否已记录。
func (d *Dedup) Seen(key string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, ok := d.seen[key]
	return ok
}

// Mark 记录已写入的键。
func (d *Dedup) Mark(key string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.seen[key] = struct{}{}
}
