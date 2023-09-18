package ptr

func New[T any](t T) *T {
	return &t
}
