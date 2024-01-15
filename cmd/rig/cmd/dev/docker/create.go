package docker

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"strings"
	"time"

	//nolint:revive
	_ "embed"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
)

func (c *Cmd) create(ctx context.Context, _ *cobra.Command, _ []string) error {
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
	}, "rig-platform-postgres", nil); err != nil {
		return err
	}

	cfg := `apiversion: config.rig.dev/v1alpha1
kind: PlatformConfig
clusters:
  docker:
    type: docker
environments:
  docker:
    cluster: docker
`

	if err := c.ensureContainer(ctx, &container.Config{
		Image: fmt.Sprint("ghcr.io/rigdev/rig-platform:", platformDockerTag),
		Env: []string{
			"RIG_CLIENT_POSTGRES_HOST=rig-platform-postgres:5432",
			"RIG_CLIENT_POSTGRES_USER=postgres",
			"RIG_CLIENT_POSTGRES_PASSWORD=postgres",
			"RIG_CLIENT_POSTGRES_INSECURE=true",
			"RIG_AUTH_SECRET=shhhdonotshare",
			"RIG_LOGGING_LEVEL=debug",
			"RIG_LOGGING_DEVMODE=true",
			"REPOSITORY_SECRET=thisisasecret",
		},
		ExposedPorts: nat.PortSet{
			"4747/tcp": struct{}{},
		},
	}, &container.HostConfig{
		PortBindings: nat.PortMap{
			"4747/tcp": []nat.PortBinding{{
				HostIP:   "127.0.0.1",
				HostPort: "4747",
			}},
		},
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: "/var/run/docker.sock",
				Target: "/var/run/docker.sock",
			},
		},
	}, "rig-platform", []byte(cfg)); err != nil {
		return err
	}

	fmt.Printf("Running init command:\n")
	initCmd := exec.CommandContext(
		ctx, "docker", "exec", "-it", "-eRIG_LOGGING_LEVEL=warn", "rig-platform", "rig-admin", "init",
	)
	initCmd.Stdin = os.Stdin
	initCmd.Stdout = os.Stdout
	initCmd.Stderr = os.Stderr

	return initCmd.Run()
}

func (c *Cmd) ensureContainer(
	ctx context.Context,
	cc *container.Config,
	chc *container.HostConfig,
	containerName string,
	config []byte,
) error {
	create := true
	_, err := c.DockerClient.ContainerInspect(ctx, containerName)
	if err != nil && !client.IsErrNotFound(err) {
		return err
	}
	if err == nil {
		ok, err := common.PromptConfirm(
			fmt.Sprint("Container `", containerName, "` already exists, re-create?"), false,
		)
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
		if err := c.ensureImage(ctx, cc.Image, strings.HasSuffix(cc.Image, ":latest")); err != nil {
			return err
		}

		fmt.Printf("Starting container `%s`... ", containerName)
		if _, err := c.DockerClient.NetworkInspect(
			ctx, "rig", types.NetworkInspectOptions{},
		); client.IsErrNotFound(err) {
			if _, err := c.DockerClient.NetworkCreate(ctx, "rig", types.NetworkCreate{
				CheckDuplicate: true,
			}); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		if _, err := c.DockerClient.ContainerCreate(ctx, cc, chc, &network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				"rig": {
					Aliases: []string{containerName},
				},
			},
		}, &v1.Platform{}, containerName); err != nil {
			return err
		}

		fmt.Printf("OK\n")
	}

	if config != nil {
		if err := c.addFile(ctx, containerName, "/etc/rig/config.yaml", config); err != nil {
			return err
		}

		if err := c.addFile(ctx, containerName, "/etc/rig/secret-config.yaml", config); err != nil {
			return err
		}
	}

	if err := c.DockerClient.ContainerStart(ctx, containerName, types.ContainerStartOptions{}); err != nil {
		return err
	}

	return nil
}

func (c *Cmd) addFile(ctx context.Context, containerName string, path string, content []byte) error {
	var buffer bytes.Buffer

	tw := tar.NewWriter(&buffer)
	defer tw.Close()

	header, err := tar.FileInfoHeader(&fileInfo{
		name: path,
		size: int64(len(content)),
	}, "")
	if err != nil {
		return err
	}

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	if _, err := tw.Write(content); err != nil {
		return err
	}

	if err := tw.Close(); err != nil {
		return err
	}

	if err := c.DockerClient.CopyToContainer(ctx, containerName, "/", bufio.NewReader(&buffer),
		types.CopyToContainerOptions{}); err != nil {
		return err
	}

	return nil
}

type fileInfo struct {
	name string
	size int64
}

func (info *fileInfo) Name() string {
	return info.name
}

func (info *fileInfo) Size() int64 {
	return info.size
}

func (info *fileInfo) IsDir() bool {
	return false
}

func (info *fileInfo) Mode() fs.FileMode {
	return 0o644
}

func (info *fileInfo) ModTime() time.Time {
	return time.Now()
}

func (info *fileInfo) Sys() any {
	return nil
}

func (c *Cmd) ensureImage(ctx context.Context, image string, force bool) error {
	if !force {
		image = strings.TrimPrefix(image, "docker.io/library/")
		image = strings.TrimPrefix(image, "index.docker.io/library/")

		images, err := c.DockerClient.ImageList(ctx, types.ImageListOptions{
			Filters: filters.NewArgs(filters.KeyValuePair{
				Key:   "reference",
				Value: image,
			}),
		})
		if err != nil {
			return err
		}

		if len(images) > 0 {
			// Image is local
			return nil
		}
	}

	fmt.Printf("Pulling image `%s`... ", image)

	r, err := c.DockerClient.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		return err
	}

	if _, err := io.Copy(io.Discard, r); err != nil {
		return err
	}

	fmt.Printf("OK\n")

	return nil
}
