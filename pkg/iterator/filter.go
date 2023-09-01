package iterator

type filter[T interface{}] struct {
	it Iterator[T]
	f  func(T) bool
}

func Filter[T interface{}](it Iterator[T], f func(T) bool) Iterator[T] {
	return &filter[T]{
		it: it,
		f:  f,
	}
}

func (f *filter[T]) Next() (T, error) {
	for {
		v, err := f.it.Next()
		if err != nil {
			return v, err
		}

		if f.f(v) {
			return v, nil
		}
	}
}

func (f *filter[T]) Close() {
	f.it.Close()
}
