package docker

import (
	"context"

	"github.com/rigdev/rig/gen/go/capsule"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/iterator"
)

func (c *Client) GetCapsuleStatus(ctx context.Context, namespace, capsuleName string) (*capsule.Status, error) {
	return nil, errors.UnimplementedErrorf("GetCapsuleStatus not implemented")
}

func (c *Client) ListInstanceStatuses(ctx context.Context, namespace, capsuleName string) (iterator.Iterator[*capsule.Instance], uint64, error) {
	return nil, 0, errors.UnimplementedErrorf("ListInstanceStatuses not implemented")
}
