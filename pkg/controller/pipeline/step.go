package pipeline

import "context"

type Step[T Request] interface {
	Apply(ctx context.Context, req T) error
}
