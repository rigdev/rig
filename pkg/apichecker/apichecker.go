package apichecker

import (
	"context"
	"fmt"

	"github.com/rigdev/rig/pkg/api/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Interface interface {
	Check(ctx context.Context) error
}

type checker struct {
	client client.Client
}

func New(client client.Client) Interface {
	return &checker{
		client: client,
	}
}

// Check implements Interface.
func (c *checker) Check(ctx context.Context) error {
	capsule := v1alpha2.Capsule{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "rig-apichecker-",
		},
		Spec: v1alpha2.CapsuleSpec{
			Image: "apichecker",
		},
	}
	if err := c.client.Create(ctx, &capsule, client.DryRunAll); err != nil {
		return fmt.Errorf("capsule create dry-run failed: %w", err)
	}
	return nil
}
