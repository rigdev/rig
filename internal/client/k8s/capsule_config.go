package k8s

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/internal/gateway/cluster"
	"github.com/rigdev/rig/pkg/api/v1alpha1"
	"github.com/rigdev/rig/pkg/iterator"
	"go.uber.org/zap"
)

func (c *Client) CreateCapsuleConfig(ctx context.Context, cfg *v1alpha1.Capsule) error {
	if err := c.rcc.CreateCapsuleConfig(ctx, cfg); err != nil {
		return err
	}

	return c.applyCapsuleConfig(ctx, cfg.Name)
}

func (c *Client) UpdateCapsuleConfig(ctx context.Context, cfg *v1alpha1.Capsule) error {
	if err := c.rcc.UpdateCapsuleConfig(ctx, cfg); err != nil {
		return err
	}

	return c.applyCapsuleConfig(ctx, cfg.Name)
}

func (c *Client) ListCapsuleConfigs(ctx context.Context, pagination *model.Pagination) (iterator.Iterator[*v1alpha1.Capsule], int64, error) {
	return c.rcc.ListCapsuleConfigs(ctx, pagination)
}

func (c *Client) applyCapsuleConfig(ctx context.Context, capsuleName string) error {
	c.logger.Debug("creating docker capsule", zap.String("capsuleName", capsuleName))

	cfg, err := c.rcc.GetCapsuleConfig(ctx, capsuleName)
	if err != nil {
		return err
	}

	envs, err := c.rcc.GetEnvironmentVariables(ctx, capsuleName)
	if err != nil {
		return err
	}

	if cfg.Spec.Image == "" {
		return nil
	}

	var auth *cluster.RegistryAuth
	// if cfg.GetRegistryAuth() != nil {
	// 	auth = &cluster.RegistryAuth{
	// 		Host: cfg.GetRegistryAuth().GetHost(),
	// 		RegistrySecret: &registry.Secret{
	// 			Username: cfg.GetRegistryAuth().GetUsername(),
	// 			Password: cfg.GetRegistryAuth().GetPassword(),
	// 		},
	// 	}
	// }

	network := &capsule.Network{}
	for _, i := range cfg.Spec.Interfaces {
		netIf := &capsule.Interface{
			Name:   i.Name,
			Port:   uint32(i.Port),
			Public: &capsule.PublicInterface{},
		}

		if i.Public != nil {
			netIf.Public.Enabled = true
			switch {
			case i.Public.Ingress != nil:
				netIf.Public.Method = &capsule.RoutingMethod{
					Kind: &capsule.RoutingMethod_Ingress_{
						Ingress: &capsule.RoutingMethod_Ingress{
							Host: i.Public.Ingress.Host,
						},
					},
				}
			case i.Public.LoadBalancer != nil:
				netIf.Public.Method = &capsule.RoutingMethod{
					Kind: &capsule.RoutingMethod_LoadBalancer_{
						LoadBalancer: &capsule.RoutingMethod_LoadBalancer{
							Port: uint32(i.Public.LoadBalancer.Port),
						},
					},
				}
			}
		}

		network.Interfaces = append(network.Interfaces, netIf)
	}

	return c.upsertCapsule(ctx, cfg.GetName(), &cluster.Capsule{
		CapsuleID: cfg.GetName(),
		Image:     cfg.Spec.Image,
		ContainerSettings: &capsule.ContainerSettings{
			EnvironmentVariables: envs,
		},
		Network:      network,
		Replicas:     uint32(cfg.Spec.Replicas),
		Namespace:    cfg.GetNamespace(),
		RegistryAuth: auth,
	})
}

func (c *Client) GetCapsuleConfig(ctx context.Context, capsuleName string) (*v1alpha1.Capsule, error) {
	return c.rcc.GetCapsuleConfig(ctx, capsuleName)
}

func (c *Client) DeleteCapsuleConfig(ctx context.Context, capsuleName string) error {
	if err := c.deleteCapsule(ctx, capsuleName); err != nil {
		return err
	}

	return c.rcc.DeleteCapsuleConfig(ctx, capsuleName)
}

func (c *Client) SetEnvironmentVariables(ctx context.Context, capsuleName string, envs map[string]string) error {
	if err := c.rcc.SetEnvironmentVariables(ctx, capsuleName, envs); err != nil {
		return err
	}

	return c.applyCapsuleConfig(ctx, capsuleName)
}

func (c *Client) GetEnvironmentVariables(ctx context.Context, capsuleName string) (map[string]string, error) {
	return c.rcc.GetEnvironmentVariables(ctx, capsuleName)
}

func (c *Client) SetEnvironmentVariable(ctx context.Context, capsuleName, name, value string) error {
	envs, err := c.rcc.GetEnvironmentVariables(ctx, capsuleName)
	if err != nil {
		return err
	}

	envs[name] = value

	if err := c.rcc.SetEnvironmentVariables(ctx, capsuleName, envs); err != nil {
		return err
	}

	return c.applyCapsuleConfig(ctx, capsuleName)
}

func (c *Client) GetEnvironmentVariable(ctx context.Context, capsuleName, name string) (value string, ok bool, err error) {
	envs, err := c.rcc.GetEnvironmentVariables(ctx, capsuleName)
	if err != nil {
		return "", false, err
	}

	if v, ok := envs[name]; ok {
		return v, ok, nil
	}

	return "", false, nil
}

func (c *Client) DeleteEnvironmentVariable(ctx context.Context, capsuleName, name string) error {
	envs, err := c.rcc.GetEnvironmentVariables(ctx, capsuleName)
	if err != nil {
		return err
	}

	delete(envs, name)

	if err := c.rcc.SetEnvironmentVariables(ctx, capsuleName, envs); err != nil {
		return err
	}

	return c.applyCapsuleConfig(ctx, capsuleName)
}
