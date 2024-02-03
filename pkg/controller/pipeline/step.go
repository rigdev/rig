package pipeline

import "context"

type Step interface {
	Apply(ctx context.Context, req Request) error
}
