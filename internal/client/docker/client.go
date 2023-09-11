package docker

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/internal/config"
	"github.com/rigdev/rig/internal/gateway/cluster"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/iterator"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Client struct {
	logger *zap.Logger
	dc     *client.Client
}

func New(cfg config.Config, logger *zap.Logger) (*Client, error) {
	var opts []client.Opt
	if cfg.Client.Docker.Host != "" {
		opts = append(opts, client.WithHost(cfg.Client.Docker.Host))
	}
	opts = append(opts, client.WithAPIVersionNegotiation())
	dc, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, err
	}

	return &Client{
		logger: logger,
		dc:     dc,
	}, nil
}

func (c *Client) Logs(ctx context.Context, capsuleName string, instanceID string, follow bool) (iterator.Iterator[*capsule.Log], error) {
	c.logger.Debug("reading docker logs", zap.String("deployment_id", capsuleName), zap.String("instance_id", instanceID))

	ls, err := c.dc.ContainerLogs(ctx, instanceID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     follow,
		Timestamps: true,
	})
	if err != nil {
		return nil, err
	}

	p := iterator.NewProducer[*capsule.Log]()

	stdout := newLogsWriter(p, false)
	stderr := newLogsWriter(p, true)
	go func() {
		_, err := stdcopy.StdCopy(stdout, stderr, ls)
		p.Error(err)
	}()

	return p, nil
}

type logsWriter struct {
	p      *iterator.Producer[*capsule.Log]
	stderr bool
}

func newLogsWriter(p *iterator.Producer[*capsule.Log], stderr bool) *logsWriter {
	return &logsWriter{
		p:      p,
		stderr: stderr,
	}
}

func (w *logsWriter) Write(bs []byte) (int, error) {
	l := &capsule.Log{
		Message: &capsule.LogMessage{},
	}
	index := bytes.IndexByte(bs, ' ')
	if index > 0 {
		if ts, err := time.Parse(time.RFC3339Nano, string(bs[:index])); err == nil {
			l.Timestamp = timestamppb.New(ts)
		}
	}

	// Note that when returning from `Write`, the buffer may no longer be referenced -> dup.
	out := bytes.Clone(bs[index+1:])
	if w.stderr {
		l.Message.Message = &capsule.LogMessage_Stderr{Stderr: out}
	} else {
		l.Message.Message = &capsule.LogMessage_Stdout{Stdout: out}
	}
	if err := w.p.Value(l); err != nil {
		return 0, err
	}

	return len(bs), nil
}

func (c *Client) createAndStartContainer(ctx context.Context, containerID string, cc *container.Config, hc *container.HostConfig, nc *network.NetworkingConfig) error {
	id, err := c.lookupContainer(ctx, containerID)
	if errors.IsNotFound(err) {
		// Already ready to create.
	} else if err != nil {
		return err
	} else {
		if err := c.dc.ContainerRemove(ctx, id, types.ContainerRemoveOptions{
			Force: true,
		}); err != nil {
			return err
		}
	}

	if _, err := c.dc.ContainerCreate(ctx, cc, hc, nc, &v1.Platform{}, containerID); err != nil {
		return err
	}

	if err := c.dc.ContainerStart(ctx, containerID, types.ContainerStartOptions{}); err != nil {
		c.logger.Info("error starting container", zap.Error(err), zap.String("instance_id", containerID))
	}

	if err := c.dc.NetworkConnect(ctx, "rig", containerID, &network.EndpointSettings{}); err != nil {
		c.logger.Warn("error adding container to rig network, RIG_HOST may not be available", zap.Error(err), zap.String("instance_id", containerID))
	}

	return nil
}

func (c *Client) lookupContainer(ctx context.Context, containerID string) (string, error) {
	c.logger.Debug("looking up docker container", zap.String("container_id", containerID))

	args := filters.NewArgs()
	args.Add("name", containerID)

	cs, err := c.dc.ContainerList(ctx, types.ContainerListOptions{
		Filters: args,
		All:     true,
	})
	if err != nil {
		return "", err
	}

	if len(cs) == 0 {
		return "", errors.NotFoundErrorf("container '%v' not found", containerID)
	}

	return cs[0].ID, err
}

func (c *Client) ensureNetwork(ctx context.Context) (string, error) {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return "", err
	}

	if _, err := c.dc.NetworkInspect(ctx, projectID.String(), types.NetworkInspectOptions{}); client.IsErrNotFound(err) {
		if _, err := c.dc.NetworkCreate(ctx, projectID.String(), types.NetworkCreate{
			CheckDuplicate: true,
		}); err != nil {
			return "", err
		}
	} else if err != nil {
		return "", err
	}

	return projectID.String(), nil
}

func (c *Client) ensureImage(ctx context.Context, image string, auth *cluster.RegistryAuth) (string, error) {
	if strings.IndexByte(image, ':') < 0 {
		image += ":latest"
	}

	is, err := c.dc.ImageList(ctx, types.ImageListOptions{
		Filters: filters.NewArgs(filters.KeyValuePair{
			Key:   "reference",
			Value: image,
		}),
	})
	if err != nil {
		return "", err
	}

	if len(is) != 0 {
		return image, nil
	}

	c.logger.Debug("pulling image", zap.String("image", image))

	opts := types.ImagePullOptions{}

	if auth != nil {
		ac := registry.AuthConfig{
			ServerAddress: auth.Host,
			Username:      auth.RegistrySecret.GetUsername(),
			Password:      auth.RegistrySecret.GetPassword(),
			Auth: base64.StdEncoding.EncodeToString(
				[]byte(fmt.Sprint(
					auth.RegistrySecret.GetUsername(),
					":",
					auth.RegistrySecret.GetPassword()),
				),
			),
		}
		secret, err := json.Marshal(ac)
		if err != nil {
			return "", err
		}

		opts.RegistryAuth = base64.StdEncoding.EncodeToString(secret)
	}

	r, err := c.dc.ImagePull(ctx, image, opts)
	if err != nil {
		return "", err
	}

	if _, err := io.Copy(io.Discard, r); err != nil {
		return "", err
	}

	return image, nil
}

func (c *Client) getContainers(ctx context.Context, prefix string) ([]types.Container, error) {
	c.logger.Debug("looking up containers", zap.String("prefix", prefix))

	args := filters.NewArgs()
	args.Add("name", fmt.Sprint(prefix, "*"))

	cs, err := c.dc.ContainerList(ctx, types.ContainerListOptions{
		Filters: args,
		All:     true,
	})
	if err != nil {
		return nil, err
	}

	return cs, nil
}

func (c *Client) ImageExistsNatively(ctx context.Context, image string) (bool, string, error) {
	fmt.Println("checking", image)
	is, err := c.dc.ImageList(ctx, types.ImageListOptions{
		Filters: filters.NewArgs(filters.KeyValuePair{
			Key:   "reference",
			Value: image,
		}),
	})
	if err != nil {
		return false, "", err
	}

	if len(is) == 0 {
		return false, "", nil
	}

	return true, is[0].ID, nil
}
