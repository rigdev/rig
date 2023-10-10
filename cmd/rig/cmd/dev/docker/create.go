package docker

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"os/exec"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
)

func (c Cmd) create(cmd *cobra.Command, args []string) error {
	ctx := c.Ctx

	v, err := c.DockerClient.VolumeCreate(ctx, volume.CreateOptions{
		Name: "rig-platform-postgres-data",
	})
	if err != nil {
		return err
	}

	if err := c.ensureContainer(ctx, &container.Config{
		Image: "postgres:latest",
		Env: []string{
			"POSTGRES_DB=rig",
			"POSTGRES_USER=postgres",
			"POSTGRES_PASSWORD=postgres",
		},
	}, &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeVolume,
				Source: v.Name,
				Target: "/var/lib/postgresql/data",
			},
		},
	}, "rig-platform-postgres"); err != nil {
		return err
	}

	if err := c.ensureContainer(ctx, &container.Config{
		Image: fmt.Sprint("ghcr.io/rigdev/rig-platform:", platformDockerTag),
		Env: []string{
			"RIG_CLIENT_POSTGRES_HOST=rig-platform-postgres:5432",
			"RIG_CLIENT_POSTGRES_USER=postgres",
			"RIG_CLIENT_POSTGRES_PASSWORD=postgres",
			"RIG_CLIENT_POSTGRES_INSECURE=true",
			"RIG_AUTH_JWT_SECRET=shhhdonotshare",
			"REPOSITORY_SECRET_POSTGRES_KEY=thisisasecret",
		},
	}, &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: "/var/run/docker.sock",
				Target: "/var/run/docker.sock",
			},
		},
	}, "rig-platform"); err != nil {
		return err
	}

	fmt.Printf("Running init command:\n")
	initCmd := exec.CommandContext(ctx, "docker", "exec", "-it", "-eRIG_LOGGING_LEVEL=warn", "rig-platform", "rig-admin", "init")
	initCmd.Stdin = os.Stdin
	initCmd.Stdout = os.Stdout
	initCmd.Stderr = os.Stderr

	return initCmd.Run()
}

func (c Cmd) ensureContainer(ctx context.Context, cc *container.Config, chc *container.HostConfig, containerName string) error {
	create := true
	if _, err := c.DockerClient.ContainerInspect(ctx, containerName); client.IsErrNotFound(err) {
	} else if err != nil {
		return err
	} else {
		ok, err := common.PromptConfirm(fmt.Sprint("Container `", containerName, "` already exists, re-create?"), false)
		if err != nil {
			return err
		}

		if ok {
			if err := c.DockerClient.ContainerRemove(ctx, containerName, types.ContainerRemoveOptions{
				Force: true,
			}); err != nil {
				return err
			}
		} else {
			create = false
		}
	}

	if create {
		fmt.Printf("Starting container `%s`... ", containerName)
		if _, err := c.DockerClient.NetworkInspect(ctx, "rig", types.NetworkInspectOptions{}); client.IsErrNotFound(err) {
			if _, err := c.DockerClient.NetworkCreate(ctx, "rig", types.NetworkCreate{
				CheckDuplicate: true,
			}); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		_, err := c.DockerClient.ContainerCreate(ctx, cc, chc, &network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				"rig": {
					Aliases: []string{containerName},
				},
			},
		}, &v1.Platform{}, containerName)
		if err != nil {
			return err
		}

		fmt.Printf("OK\n")
	}

	if err := c.DockerClient.ContainerStart(ctx, containerName, types.ContainerStartOptions{}); err != nil {
		return err
	}

	return nil
}
