package iterator

import (
	"context"
	"io"
	"sync"
)

type Producer[T interface{}] struct {
	mutex  sync.Mutex
	ctx    context.Context
	cancel context.CancelFunc
	c      chan T
	cErr   chan struct{}
	err    error
}

func NewProducer[T interface{}]() *Producer[T] {
	return NewBufferedProducer[T](0)
}

func NewBufferedProducer[T interface{}](bufferSize int) *Producer[T] {
	ctx, cancel := context.WithCancel(context.Background())
	return &Producer[T]{
		ctx:    ctx,
		cancel: cancel,
		c:      make(chan T, bufferSize),
		cErr:   make(chan struct{}),
	}
}

func (p *Producer[T]) Next() (T, error) {
	select {
	case <-p.cErr:
		p.mutex.Lock()
		defer p.mutex.Unlock()
		var t T
		return t, p.err
	default:
	}

	select {
	case <-p.cErr:
		p.mutex.Lock()
		defer p.mutex.Unlock()
		var t T
		return t, p.err
	case <-p.ctx.Done():
		var t T
		return t, p.ctx.Err()
	case t := <-p.c:
		return t, nil
	}
}

func (p *Producer[T]) Close() {
	p.cancel()
}

// Value adds a new value to the iterator. This method will block
// until the consumer has read out the value or the iterator is closed.
// The returning error should be read, as an error indicates no more values
// can be written to the producer.
func (p *Producer[T]) Value(t T) error {
	select {
	case <-p.ctx.Done():
		return p.ctx.Err()
	case p.c <- t:
		return nil
	}
}

// Error closes the producer with the provided error. If the error is `nil`,
// io.EOF will be used instead to indicate a nice close.
func (p *Producer[T]) Error(err error) {
	if err == nil {
		err = io.EOF
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()
	if p.err == nil {
		p.err = err
		close(p.cErr)
	}
}

func (p *Producer[T]) Done() {
	p.Error(nil)
}
