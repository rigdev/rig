package k8s

import (
	"context"

	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/gen/go/capsule"
	"github.com/rigdev/rig/gen/go/registry"
	"github.com/rigdev/rig/internal/gateway/cluster"
	"github.com/rigdev/rig/pkg/iterator"
	"go.uber.org/zap"
)

func (c *Client) CreateCapsuleConfig(ctx context.Context, cfg *capsule.Config) error {
	if err := c.rcc.CreateCapsuleConfig(ctx, cfg); err != nil {
		return err
	}

	return c.applyCapsuleConfig(ctx, cfg)
}

func (c *Client) UpdateCapsuleConfig(ctx context.Context, cfg *capsule.Config) error {
	if err := c.rcc.UpdateCapsuleConfig(ctx, cfg); err != nil {
		return err
	}

	return c.applyCapsuleConfig(ctx, cfg)
}

func (c *Client) ListCapsuleConfigs(ctx context.Context, pagination *model.Pagination) (iterator.Iterator[*capsule.Config], int64, error) {
	return c.rcc.ListCapsuleConfigs(ctx, pagination)
}

func (c *Client) applyCapsuleConfig(ctx context.Context, cfg *capsule.Config) error {
	c.logger.Debug("creating docker capsule", zap.String("capsuleName", cfg.GetName()))

	if cfg.GetImage() == "" {
		return nil
	}

	var auth *cluster.RegistryAuth
	if cfg.GetRegistryAuth() != nil {
		auth = &cluster.RegistryAuth{
			Host: cfg.GetRegistryAuth().GetHost(),
			RegistrySecret: &registry.Secret{
				Username: cfg.GetRegistryAuth().GetUsername(),
				Password: cfg.GetRegistryAuth().GetPassword(),
			},
		}
	}

	return c.upsertCapsule(ctx, cfg.GetName(), &cluster.Capsule{
		CapsuleID:         cfg.GetName(),
		Image:             cfg.GetImage(),
		ContainerSettings: cfg.GetContainerSettings(),
		Network:           cfg.GetNetwork(),
		Replicas:          cfg.GetReplicas(),
		Namespace:         cfg.GetNamespace(),
		RegistryAuth:      auth,
	})
}

func (c *Client) GetCapsuleConfig(ctx context.Context, capsuleName string) (*capsule.Config, error) {
	return c.rcc.GetCapsuleConfig(ctx, capsuleName)
}

func (c *Client) DeleteCapsuleConfig(ctx context.Context, capsuleName string) error {
	if err := c.deleteCapsule(ctx, capsuleName); err != nil {
		return err
	}

	return c.rcc.DeleteCapsuleConfig(ctx, capsuleName)
}
