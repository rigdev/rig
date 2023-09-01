package iterator

import "io"

type fromList[T interface{}] struct {
	ts    []T
	index int
}

func FromList[T interface{}](ts []T) Iterator[T] {
	return &fromList[T]{
		ts: ts,
	}
}

func (l *fromList[T]) Next() (T, error) {
	if l.index >= len(l.ts) {
		var t T
		return t, io.EOF
	}

	i := l.index
	l.index++
	return l.ts[i], nil
}

func (l *fromList[T]) Close() {
	l.index = len(l.ts)
}
