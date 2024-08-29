package ptr

import "golang.org/x/exp/constraints"

func New[T any](t T) *T {
	return &t
}

type numeric interface {
	constraints.Integer | constraints.Float
}

func Convert[T, V numeric](t *T) *V {
	if t == nil {
		return nil
	}
	tt := *t
	v := (V)(tt)
	return &v
}

func Copy[T any](t *T) *T {
	if t == nil {
		return nil
	}
	tt := *t
	return &tt
}
