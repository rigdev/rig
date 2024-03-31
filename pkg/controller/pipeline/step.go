package pipeline

import "context"

type CapsuleStep interface {
	Apply(ctx context.Context, req CapsuleRequest) error
}
