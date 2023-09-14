package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	api_capsule "github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/gen/go/capsule"
	"github.com/rigdev/rig/gen/go/registry"
	"github.com/rigdev/rig/internal/gateway/cluster"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/iterator"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
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

	netID, err := c.ensureNetwork(ctx)
	if err != nil {
		return err
	}

	var regAuth *cluster.RegistryAuth
	if cfg.GetRegistryAuth() != nil {
		regAuth = &cluster.RegistryAuth{
			Host: cfg.GetRegistryAuth().GetHost(),
			RegistrySecret: &registry.Secret{
				Username: cfg.GetRegistryAuth().GetUsername(),
				Password: cfg.GetRegistryAuth().GetPassword(),
			},
		}
	}

	image := cfg.GetImage()
	if err := c.ensureImage(ctx, image, regAuth); err != nil {
		return err
	}

	var cmd []string
	if cfg.GetContainerSettings().GetCommand() != "" {
		cmd = append([]string{cfg.GetContainerSettings().GetCommand()}, cfg.GetContainerSettings().GetArgs()...)
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
	for k, v := range cfg.GetContainerSettings().GetEnvironmentVariables() {
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

	for _, e := range cfg.GetNetwork().GetInterfaces() {
		if e.GetPublic().GetEnabled() {
			switch v := e.GetPublic().GetMethod().GetKind().(type) {
			case *api_capsule.RoutingMethod_LoadBalancer_:
				dcc.ExposedPorts[nat.Port(fmt.Sprint(v.LoadBalancer.GetPort(), "/tcp"))] = struct{}{}
			default:
				return errors.InvalidArgumentErrorf("docker only supports LoadBalancer as routing method for public interfaces")
			}
		}
	}

	existing, err := c.getInstances(ctx, cfg.GetName())
	if err != nil {
		return err
	}

	for i := 0; i < int(cfg.GetReplicas()); i++ {
		containerID := fmt.Sprint(cfg.GetName(), "-instance-", i)

		dnc.EndpointsConfig[netID] = &network.EndpointSettings{
			Aliases: []string{cfg.GetName(), containerID},
		}
		if err := c.deleteService(ctx, cfg.GetName()); err != nil {
			return err
		}

		if err := c.createAndStartContainer(ctx, containerID, dcc, dhc, dnc, cfg.GetConfigFiles()); err != nil {
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

func (c *Client) GetCapsuleConfig(ctx context.Context, capsuleName string) (*capsule.Config, error) {
	return c.rcc.GetCapsuleConfig(ctx, capsuleName)
}

func (c *Client) DeleteCapsuleConfig(ctx context.Context, capsuleName string) error {
	cs, err := c.getInstances(ctx, capsuleName)
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

	return c.rcc.DeleteCapsuleConfig(ctx, capsuleName)
}
