package docker

import (
	"context"
	"fmt"
	"strconv"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/rigdev/rig/gen/go/proxy"
	"github.com/rigdev/rig/internal/build"
	"google.golang.org/protobuf/encoding/protojson"
)

func (c *Client) upsertService(ctx context.Context, capsuleName string, pc *proxy.Config) error {
	containerID := fmt.Sprint(capsuleName, "-service")

	bs, err := protojson.Marshal(pc)
	if err != nil {
		return err
	}

	cfg := strconv.QuoteToASCII(string(bs))

	image, err := c.ensureImage(ctx, fmt.Sprint("ghcr.io/rigdev/rig:", build.Version()), nil)
	if err != nil {
		return err
	}

	cc := &container.Config{
		Image:        image,
		Cmd:          []string{"rig-proxy"},
		ExposedPorts: nat.PortSet{},
		Volumes:      map[string]struct{}{},
		Env: []string{
			fmt.Sprint("RIG_PROXY_CONFIG=", cfg),
		},
	}

	netID, err := c.ensureNetwork(ctx)
	if err != nil {
		return err
	}

	hc := &container.HostConfig{
		NetworkMode:  container.NetworkMode(netID),
		PortBindings: nat.PortMap{},
	}

	nc := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			netID: {
				Aliases: []string{capsuleName, fmt.Sprint(capsuleName, ".local")},
			},
		},
	}

	for _, e := range pc.GetInterfaces() {
		cc.ExposedPorts[nat.Port(fmt.Sprint(e.GetSourcePort(), "/tcp"))] = struct{}{}
		hc.PortBindings[nat.Port(fmt.Sprint(e.GetSourcePort(), "/tcp"))] = []nat.PortBinding{{
			HostIP:   "127.0.0.1",
			HostPort: fmt.Sprint(e.GetSourcePort()),
		}}
	}

	if err := c.createAndStartContainer(ctx, containerID, cc, hc, nc); err != nil {
		return err
	}

	return nil
}

func (c *Client) deleteService(ctx context.Context, capsuleName string) error {
	containerID := fmt.Sprint(capsuleName, "-service")
	if err := c.dc.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{
		Force: true,
	}); client.IsErrNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	return nil
}
