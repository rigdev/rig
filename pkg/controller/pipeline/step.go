package pipeline

import "context"

type Step interface {
	Apply(ctx context.Context, req CapsuleRequest) error
}
