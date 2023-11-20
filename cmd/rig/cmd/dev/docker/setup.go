package docker

import (
	"github.com/docker/docker/client"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var platformDockerTag string

type Cmd struct {
	fx.In

	DockerClient *client.Client
}

var cmd Cmd

func initCmd(c Cmd) {
	cmd.DockerClient = c.DockerClient
}

func Setup(parent *cobra.Command) {
	docker := &cobra.Command{
		Use:               "docker",
		Short:             "The docker command is used to setup and manage a development Docker cluster running Rig",
		PersistentPreRunE: base.MakeInvokePreRunE(initCmd),
	}

	create := &cobra.Command{
		Use:   "create",
		Short: "Create a Rig cluster in Docker for local development",
		Args:  cobra.NoArgs,
		RunE:  base.CtxWrap(cmd.create),
		Annotations: map[string]string{
			base.OmitUser:    "",
			base.OmitProject: "",
		},
	}
	create.Flags().StringVarP(
		&platformDockerTag,
		"platform-docker-tag", "p", "latest", "The rig-platform docker image tag. Defaults to latest.",
	)
	docker.AddCommand(create)

	parent.AddCommand(docker)
}
