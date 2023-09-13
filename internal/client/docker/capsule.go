package docker

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/internal/gateway/cluster"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/iterator"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	_rigCapsuleIDLabel = "io.rig.capsule-id"
	_rigProjectIDLabel = "io.rig.project-id"
)

func (c *Client) upsertCapsule(ctx context.Context, capsuleName string, cc *cluster.Capsule) error {
	c.logger.Debug("creating docker capsule", zap.String("capsuleName", capsuleName))

	netID, err := c.ensureNetwork(ctx)
	if err != nil {
		return err
	}

	if err := c.ensureImage(ctx, cc.Image, cc.RegistryAuth); err != nil {
		return err
	}

	var cmd []string
	if cc.ContainerSettings.GetCommand() != "" {
		cmd = append([]string{cc.ContainerSettings.GetCommand()}, cc.ContainerSettings.GetArgs()...)
	}

	pid, err := auth.GetProjectID(ctx)
	if err != nil {
		return err
	}

	dcc := &container.Config{
		Image:        cc.Image,
		Cmd:          cmd,
		ExposedPorts: nat.PortSet{},
		Labels: map[string]string{
			_rigCapsuleIDLabel: cc.CapsuleID,
			_rigProjectIDLabel: pid.String(),
		},
		Env: []string{
			// TODO(anders): Get port from config.
			"RIG_HOST=http://rig:4747",
		},
	}
	for k, v := range cc.ContainerSettings.GetEnvironmentVariables() {
		dcc.Env = append(dcc.Env, fmt.Sprint(k, "=", v))
	}

	limits := cc.ContainerSettings.GetResources().GetLimits()
	dhc := &container.HostConfig{
		NetworkMode:  container.NetworkMode(netID),
		PortBindings: nat.PortMap{},
		RestartPolicy: container.RestartPolicy{
			Name: "always",
		},
		Resources: container.Resources{
			Memory:   int64(limits.GetMemoryBytes()),
			NanoCPUs: int64(limits.GetCpuMillis() * 1_000_000),
		},
	}

	dnc := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{},
	}

	useProxy := true
	for _, e := range cc.Network.GetInterfaces() {
		if e.GetPublic().GetEnabled() {
			switch v := e.GetPublic().GetMethod().GetKind().(type) {
			case *capsule.RoutingMethod_LoadBalancer_:
				dcc.ExposedPorts[nat.Port(fmt.Sprint(v.LoadBalancer.GetPort(), "/tcp"))] = struct{}{}
			default:
				return errors.InvalidArgumentErrorf("docker only supports LoadBalancer as routing method for public interfaces")
			}
		}
	}

	pc, err := cluster.CreateProxyConfig(ctx, cc.Network, cc.JWTMethod)
	if err != nil {
		return err
	}

	existing, err := c.getInstances(ctx, capsuleName)
	if err != nil {
		return err
	}

	for i := 0; i < int(cc.Replicas); i++ {
		containerID := fmt.Sprint(capsuleName, "-instance-", i)

		if useProxy {
			pc.TargetHost = fmt.Sprint("instances.", capsuleName)
			dnc.EndpointsConfig[netID] = &network.EndpointSettings{
				Aliases: []string{
					pc.GetTargetHost(),
					containerID,
				},
			}
			if err := c.upsertService(ctx, capsuleName, pc); err != nil {
				return err
			}
		} else {
			dnc.EndpointsConfig[netID] = &network.EndpointSettings{
				Aliases: []string{capsuleName, containerID},
			}
			if err := c.deleteService(ctx, capsuleName); err != nil {
				return err
			}
		}

		dhc.Mounts = nil
		for v, p := range cc.Volumes {
			if !strings.Contains(v, "/") {
				v = fmt.Sprint(v, "-", i)
				if err := c.CreateVolume(ctx, v); err != nil {
					return err
				}
				dhc.Mounts = append(dhc.Mounts, mount.Mount{
					Type:   mount.TypeVolume,
					Source: v,
					Target: p,
				})
			} else {
				dhc.Mounts = append(dhc.Mounts, mount.Mount{
					Type:   mount.TypeBind,
					Source: v,
					Target: p,
				})
			}
		}

		err := c.createAndStartContainer(
			ctx,
			containerID,
			dcc,
			dhc,
			dnc,
			cc.ConfigFiles,
		)
		if err != nil {
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

func (c *Client) ListInstances(ctx context.Context, capsuleName string) (iterator.Iterator[*capsule.Instance], uint64, error) {
	c.logger.Debug("looking up capsule instances", zap.String("capsule_name", capsuleName))

	cs, err := c.getInstances(ctx, capsuleName)
	if err != nil {
		return nil, 0, err
	}

	var is []*capsule.Instance
	for _, ci := range cs {
		i := &capsule.Instance{
			InstanceId: containerName(ci),
			CreatedAt:  timestamppb.New(time.Unix(ci.Created, 0)),
		}

		cj, err := c.dc.ContainerInspect(ctx, ci.ID)
		if client.IsErrNotFound(err) {
			continue
		} else if err != nil {
			return nil, 0, err
		}

		i.RestartCount = uint32(cj.RestartCount)

		i.BuildId = ci.Image

		if cj.State.StartedAt != "" {
			if sa, err := time.Parse(time.RFC3339Nano, cj.State.StartedAt); err != nil {
				c.logger.Debug("invalid started at", zap.Error(err), zap.String("started_at", cj.State.StartedAt))
			} else {
				i.StartedAt = timestamppb.New(sa)
			}
		}

		if cj.State.FinishedAt != "" {
			if fa, err := time.Parse(time.RFC3339Nano, cj.State.FinishedAt); err != nil {
				c.logger.Debug("invalid finished at", zap.Error(err), zap.String("finished_at", cj.State.FinishedAt))
			} else {
				i.FinishedAt = timestamppb.New(fa)
			}
		}

		switch ci.State {
		case "exited":
			if cj.State.ExitCode == 0 {
				i.State = capsule.State_STATE_SUCCEEDED
			} else {
				i.State = capsule.State_STATE_FAILED
			}
		case "running":
			i.State = capsule.State_STATE_RUNNING
		case "created":
			i.State = capsule.State_STATE_PENDING
		default:
			c.logger.Warn("invalid container state", zap.String("state", ci.State))
		}

		is = append(is, i)
	}

	return iterator.FromList(is), uint64(len(is)), nil
}

func (c *Client) RestartInstance(ctx context.Context, deploymentID, instanceID string) error {
	if err := c.dc.ContainerRestart(ctx, instanceID, container.StopOptions{}); client.IsErrNotFound(err) {
		return errors.NotFoundErrorf("instance '%v' not found", instanceID)
	} else if err != nil {
		return err
	}

	return nil
}

func (c *Client) deleteCapsule(ctx context.Context, capsuleName string) error {
	c.logger.Debug("delete docker capsule", zap.String("capsule_name", capsuleName))

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

	return nil
}

func (c *Client) getInstances(ctx context.Context, capsuleName string) ([]types.Container, error) {
	c.logger.Debug("looking up capsule containers", zap.String("capsule_name", capsuleName))

	return c.getContainers(ctx, fmt.Sprint(capsuleName, "-instance-"))
}

func containerName(c types.Container) string {
	return strings.TrimPrefix(c.Names[0], "/")
}
