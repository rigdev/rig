package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/internal/gateway/cluster"
	"github.com/rigdev/rig/pkg/api/v1alpha1"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/iterator"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
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

func (c *Client) applyCapsuleConfig(ctx context.Context, capsuleID string) error {
	c.logger.Debug("creating docker capsule", zap.String("capsuleID", capsuleID))

	cfg, err := c.rcc.GetCapsuleConfig(ctx, capsuleID)
	if err != nil {
		return err
	}

	envs, err := c.rcc.GetEnvironmentVariables(ctx, capsuleID)
	if err != nil {
		return err
	}

	image := cfg.Spec.Image
	if image == "" {
		return nil
	}

	netID, err := c.ensureNetwork(ctx)
	if err != nil {
		return err
	}

	var regAuth *cluster.RegistryAuth
	// if cfg.GetRegistryAuth() != nil {
	// 	regAuth = &cluster.RegistryAuth{
	// 		Host: cfg.GetRegistryAuth().GetHost(),
	// 		RegistrySecret: &registry.Secret{
	// 			Username: cfg.GetRegistryAuth().GetUsername(),
	// 			Password: cfg.GetRegistryAuth().GetPassword(),
	// 		},
	// 	}
	// }

	if err := c.ensureImage(ctx, image, regAuth); err != nil {
		return err
	}

	var cmd []string
	if cfg.Spec.Command != "" {
		cmd = append([]string{cfg.Spec.Image}, cfg.Spec.Args...)
	}

	dcc := &container.Config{
		Image:        image,
		Cmd:          cmd,
		ExposedPorts: nat.PortSet{},
		Labels: map[string]string{
			_rigCapsuleIDLabel: cfg.GetName(),
			_rigProjectIDLabel: cfg.GetNamespace(),
		},
		Env: []string{
			// TODO(anders): Get port from config.
			"RIG_HOST=http://rig:4747",
		},
	}
	for k, v := range envs {
		dcc.Env = append(dcc.Env, fmt.Sprint(k, "=", v))
	}

	dhc := &container.HostConfig{
		NetworkMode:  container.NetworkMode(netID),
		PortBindings: nat.PortMap{},
		RestartPolicy: container.RestartPolicy{
			Name: "always",
		},
	}

	dnc := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{},
	}

	for _, i := range cfg.Spec.Interfaces {
		if i.Public != nil {
			switch {
			case i.Public.LoadBalancer != nil:
				dcc.ExposedPorts[nat.Port(fmt.Sprint(i.Public.LoadBalancer.Port, "/tcp"))] = struct{}{}
			default:
				return errors.InvalidArgumentErrorf("docker only supports LoadBalancer as routing method for public interfaces")
			}
		}
	}

	existing, err := c.getInstances(ctx, cfg.GetName())
	if err != nil {
		return err
	}

	for i := 0; i < int(cfg.Spec.Replicas); i++ {
		containerID := fmt.Sprint(cfg.GetName(), "-instance-", i)

		dnc.EndpointsConfig[netID] = &network.EndpointSettings{
			Aliases: []string{cfg.GetName(), containerID},
		}
		if err := c.deleteService(ctx, cfg.GetName()); err != nil {
			return err
		}

		if err := c.createAndStartContainer(ctx, containerID, dcc, dhc, dnc, cfg.Spec.Files); err != nil {
			return err
		}

		if i := slices.IndexFunc(existing, func(c types.Container) bool { return containerName(c) == containerID }); i >= 0 {
			existing = slices.Delete(existing, i, i+1)
		}
	}

	for _, e := range existing {
		if err := c.dc.ContainerRemove(ctx, containerName(e), types.ContainerRemoveOptions{
			Force: true,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) GetCapsuleConfig(ctx context.Context, capsuleID string) (*v1alpha1.Capsule, error) {
	return c.rcc.GetCapsuleConfig(ctx, capsuleID)
}

func (c *Client) DeleteCapsuleConfig(ctx context.Context, capsuleID string) error {
	cs, err := c.getInstances(ctx, capsuleID)
	if err != nil {
		return err
	}

	for _, ci := range cs {
		if err := c.dc.ContainerRemove(ctx, containerName(ci), types.ContainerRemoveOptions{
			Force: true,
		}); err != nil {
			return err
		}
	}

	return c.rcc.DeleteCapsuleConfig(ctx, capsuleID)
}

func (c *Client) SetEnvironmentVariables(ctx context.Context, capsuleID string, envs map[string]string) error {
	if err := c.rcc.SetEnvironmentVariables(ctx, capsuleID, envs); err != nil {
		return err
	}

	return c.applyCapsuleConfig(ctx, capsuleID)
}

func (c *Client) GetEnvironmentVariables(ctx context.Context, capsuleID string) (map[string]string, error) {
	return c.rcc.GetEnvironmentVariables(ctx, capsuleID)
}

func (c *Client) SetEnvironmentVariable(ctx context.Context, capsuleID, name, value string) error {
	envs, err := c.rcc.GetEnvironmentVariables(ctx, capsuleID)
	if err != nil {
		return err
	}

	envs[name] = value

	if err := c.rcc.SetEnvironmentVariables(ctx, capsuleID, envs); err != nil {
		return err
	}

	return c.applyCapsuleConfig(ctx, capsuleID)
}

func (c *Client) GetEnvironmentVariable(ctx context.Context, capsuleID, name string) (value string, ok bool, err error) {
	envs, err := c.rcc.GetEnvironmentVariables(ctx, capsuleID)
	if err != nil {
		return "", false, err
	}

	if v, ok := envs[name]; ok {
		return v, ok, nil
	}

	return "", false, nil
}

func (c *Client) DeleteEnvironmentVariable(ctx context.Context, capsuleID, name string) error {
	envs, err := c.rcc.GetEnvironmentVariables(ctx, capsuleID)
	if err != nil {
		return err
	}

	delete(envs, name)

	if err := c.rcc.SetEnvironmentVariables(ctx, capsuleID, envs); err != nil {
		return err
	}

	return c.applyCapsuleConfig(ctx, capsuleID)
}
