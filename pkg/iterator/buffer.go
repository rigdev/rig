package iterator

import "io"

type buffer[T interface{}] struct {
	it Iterator[T]
	p  *Producer[T]
}

func NewBuffer[T interface{}](it Iterator[T], bufferSize int) Iterator[T] {
	b := &buffer[T]{
		it: it,
		p:  NewBufferedProducer[T](bufferSize),
	}
	go b.run()
	return b
}

func (b *buffer[T]) Next() (T, error) {
	return b.p.Next()
}

func (b *buffer[T]) Close() {
	b.p.Close()
}

func (b *buffer[T]) run() {
	defer b.it.Close()
	for {
		t, err := b.it.Next()
		if err == io.EOF {
			b.p.Done()
			return
		} else if err != nil {
			b.p.Error(err)
			return
		}

		if err := b.p.Value(t); err != nil {
			b.p.Error(err)
			return
		}
	}
}
