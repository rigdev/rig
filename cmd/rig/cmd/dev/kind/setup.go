package kind

import (
	"context"

	"github.com/docker/docker/client"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/rigdev/rig/cmd/rig/cmd/cmd_config"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var (
	dockerTag    string
	helmChartTag string
	chartPath    string
)

type Cmd struct {
	fx.In

	Ctx          context.Context
	DockerClient *client.Client
	Rig          rig.Client
	Cfg          *cmd_config.Config
}

func (c *Cmd) Setup(parent *cobra.Command) {
	kind := &cobra.Command{
		Use:   "kind",
		Short: "The kind command is used to setup and manage a development kubernetes cluster running Rig using Kind",
	}

	create := &cobra.Command{
		Use:   "create",
		Short: "Create a rig cluster in Kind for local development",
		Args:  cobra.NoArgs,
		RunE:  c.create,
		Annotations: map[string]string{
			base.OmitUser:    "",
			base.OmitProject: "",
		},
	}
	create.Flags().StringVarP(&dockerTag, "docker-tag", "d", "", "The Rig docker image tag. Defaults to the latest one")
	create.Flags().StringVarP(&helmChartTag, "helm-chart-tag", "c", "", "The tag of Rig's Helm chart. Defaults to the latest one")
	create.Flags().StringVarP(&chartPath, "chart-path", "p", "", "If set, uses the helm chart at chart-path to build Rig.")
	kind.AddCommand(create)

	deploy := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy a new (or specific) version of Rig to the kind cluster",
		Args:  cobra.NoArgs,
		RunE:  c.deploy,
		Annotations: map[string]string{
			base.OmitUser:    "",
			base.OmitProject: "",
		},
	}
	kind.AddCommand(deploy)
	deploy.Flags().StringVarP(&dockerTag, "docker-tag", "d", "", "The Rig docker image tag. Defaults to the latest one")
	deploy.Flags().StringVarP(&helmChartTag, "helm-chart-tag", "c", "", "The tag of Rig's Helm chart. Defaults to the latest one")
	deploy.Flags().StringVarP(&chartPath, "chart-path", "p", "", "If set, uses the helm chart at chart-path to build Rig.")

	clean := &cobra.Command{
		Use:   "clean",
		Short: "Deletes the rig kind-cluster",
		Args:  cobra.NoArgs,
		RunE:  c.clean,
		Annotations: map[string]string{
			base.OmitUser:    "",
			base.OmitProject: "",
		},
	}
	kind.AddCommand(clean)

	parent.AddCommand(kind)
}
