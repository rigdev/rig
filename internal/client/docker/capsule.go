package docker

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/iterator"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	_rigCapsuleIDLabel = "io.rig.capsule-id"
	_rigProjectIDLabel = "io.rig.project-id"
)

func (c *Client) ListInstances(ctx context.Context, capsuleID string) (iterator.Iterator[*capsule.Instance], uint64, error) {
	c.logger.Debug("looking up capsule instances", zap.String("capsule_name", capsuleID))

	cs, err := c.getInstances(ctx, capsuleID)
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

func (c *Client) deleteCapsule(ctx context.Context, capsuleID string) error {
	c.logger.Debug("delete docker capsule", zap.String("capsule_name", capsuleID))

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

	return nil
}

func (c *Client) getInstances(ctx context.Context, capsuleID string) ([]types.Container, error) {
	c.logger.Debug("looking up capsule containers", zap.String("capsule_name", capsuleID))

	return c.getContainers(ctx, fmt.Sprint(capsuleID, "-instance-"))
}

func containerName(c types.Container) string {
	return strings.TrimPrefix(c.Names[0], "/")
}
