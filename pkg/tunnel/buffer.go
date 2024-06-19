package tunnel

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/rigdev/rig/pkg/errors"
)

const _maxBufferSize = 1024 * 1024

type Buffer struct {
	segments [][]byte
	total    int
	lock     sync.Mutex
	closed   bool
	notify   chan struct{}
}

func NewBuffer() *Buffer {
	return &Buffer{
		notify: make(chan struct{}, 1),
	}
}

func (b *Buffer) Put(ctx context.Context, data []byte) error {
	start := time.Now()
	for {
		b.lock.Lock()
		if b.closed {
			return io.ErrClosedPipe
		}

		if b.total == 0 || b.total+len(data) <= _maxBufferSize {
			b.segments = append(b.segments, data)
			b.total += len(data)
			select {
			case b.notify <- struct{}{}:
			default:
			}
			b.lock.Unlock()
			return nil
		}

		b.lock.Unlock()

		if time.Since(start) > 3*time.Second {
			return errors.DeadlineExceededErrorf("timeout writing to tunneling connection")
		}

		select {
		case <-time.After(10 * time.Millisecond):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (b *Buffer) Close() {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.closed = true
	select {
	case b.notify <- struct{}{}:
	default:
	}
}

func (b *Buffer) Take(ctx context.Context) ([]byte, error) {
	for {
		b.lock.Lock()
		if len(b.segments) > 0 {
			data := b.segments[0]
			b.segments = b.segments[1:]
			b.total -= len(data)
			b.lock.Unlock()
			return data, nil
		}
		if b.closed {
			b.lock.Unlock()
			return nil, io.EOF
		}

		b.lock.Unlock()

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-b.notify:
		}
	}
}
