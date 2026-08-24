package source

import (
	"context"
	"sync/atomic"
)

// Conn 表示一个源连接，release 时归还连接池。
type Conn struct {
	pool   *ConnPool
	active atomic.Bool
}

// release 归还连接，幂等。
func (c *Conn) release() error {
	if c.active.CompareAndSwap(true, false) {
		c.pool.releaseOne()
	}
	return nil
}

// ConnPool 跟踪源连接的生命周期。
type ConnPool struct {
	active atomic.Int64
	total  atomic.Int64
}

// NewConnPool 创建连接池。
func NewConnPool() *ConnPool { return &ConnPool{} }

// Acquire 分配一个连接。
func (p *ConnPool) Acquire(ctx context.Context) (*Conn, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	p.active.Add(1)
	p.total.Add(1)
	conn := &Conn{pool: p}
	conn.active.Store(true)
	return conn, nil
}

func (p *ConnPool) releaseOne() { p.active.Add(-1) }

// Active 返回当前活跃连接数。
func (p *ConnPool) Active() int64 { return p.active.Load() }

// Total 返回累计分配连接数。
func (p *ConnPool) Total() int64 { return p.total.Load() }
