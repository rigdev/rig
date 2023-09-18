package cluster

import (
	"context"

	"github.com/rigdev/rig/gen/go/capsule"
	"github.com/rigdev/rig/pkg/iterator"
)

type StatusGateway interface {
	GetCapsuleStatus(ctx context.Context, namespace, capsuleID string) (*capsule.Status, error)

	ListInstanceStatuses(ctx context.Context, namespace, capsuleID string) (iterator.Iterator[*capsule.Instance], uint64, error)
	RestartInstance(ctx context.Context, capsuleID, instanceID string) error

	// Logs(ctx context.Context, capsuleID, instanceID string, follow bool) (iterator.Iterator[*capsule.Log], error)

	// ListCapsuleMetrics(ctx context.Context) (iterator.Iterator[*capsule.InstanceMetrics], error)

	// ImageExistsNatively checks if the image exists natively in the cluster. The repo digest is returned if found.
	ImageExistsNatively(ctx context.Context, image string) (bool, string, error)
}
