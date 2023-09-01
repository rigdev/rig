package iterator

type MapFunc[F, T interface{}] func(F) (T, error)

type mapIterator[F, T interface{}] struct {
	it Iterator[F]
	f  MapFunc[F, T]
}

func Map[F, T interface{}](it Iterator[F], f MapFunc[F, T]) Iterator[T] {
	return &mapIterator[F, T]{
		it: it,
		f:  f,
	}
}

func (m *mapIterator[F, T]) Next() (T, error) {
	f, err := m.it.Next()
	if err != nil {
		var t T
		return t, err
	}

	t, err := m.f(f)
	if err != nil {
		m.it.Close()
		var t T
		return t, err
	}

	return t, nil
}

func (m *mapIterator[F, T]) Close() {
	m.it.Close()
}
