package kind

import (
	"github.com/docker/docker/client"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/rigdev/rig/cmd/rig/cmd/cmd_config"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var (
	platformDockerTag string
	platformChartPath string
	operatorDockerTag string
	operatorChartPath string
)

type Cmd struct {
	fx.In

	DockerClient *client.Client
	Rig          rig.Client
	Cfg          *cmd_config.Config
}

func Setup(parent *cobra.Command) {
	kind := &cobra.Command{
		Use:   "kind",
		Short: "The kind command is used to setup and manage a development kubernetes cluster running Rig using Kind",
	}

	create := &cobra.Command{
		Use:   "create",
		Short: "Create a rig cluster in Kind for local development",
		Args:  cobra.NoArgs,
		RunE:  base.Register(func(c Cmd) any { return c.create }),
		Annotations: map[string]string{
			base.OmitUser:    "",
			base.OmitProject: "",
		},
		ValidArgsFunction: common.NoCompletions,
	}
	create.Flags().StringVarP(&platformDockerTag, "platform-docker-tag", "p", "", "The rig-platform docker image tag. Defaults to latest.")
	create.Flags().StringVar(&platformChartPath, "platform-chart-path", "", "If set, uses the helm chart at platform-chart-path to build rig-platform.")
	create.Flags().StringVarP(&operatorDockerTag, "operator-docker-tag", "o", "", "The rig-operator docker image tag. Defaults to latest.")
	create.Flags().StringVar(&operatorChartPath, "operator-chart-path", "", "If set, uses the helm chart at operator-chart-path to build rig-operator.")
	create.RegisterFlagCompletionFunc("platform-docker-tag", common.NoCompletions)
	create.RegisterFlagCompletionFunc("platform-chart-path", common.NoCompletions)
	create.RegisterFlagCompletionFunc("operator-docker-tag", common.NoCompletions)
	create.RegisterFlagCompletionFunc("operator-chart-path", common.NoCompletions)
	kind.AddCommand(create)

	deploy := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy a new (or specific) version of Rig to the kind cluster",
		Args:  cobra.NoArgs,
		RunE:  base.Register(func(c Cmd) any { return c.deploy }),
		Annotations: map[string]string{
			base.OmitUser:    "",
			base.OmitProject: "",
		},
		ValidArgsFunction: common.NoCompletions,
	}
	kind.AddCommand(deploy)
	deploy.Flags().StringVarP(&platformDockerTag, "platform-docker-tag", "p", "", "The rig-platform docker image tag. Defaults to latest.")
	deploy.Flags().StringVar(&platformChartPath, "platform-chart-path", "", "If set, uses the helm chart at platform-chart-path to build rig-platform.")
	deploy.Flags().StringVarP(&operatorDockerTag, "operator-docker-tag", "o", "", "The rig-operator docker image tag. Defaults to latest.")
	deploy.Flags().StringVar(&operatorChartPath, "operator-chart-path", "", "If set, uses the helm chart at operator-chart-path to build rig-operator.")
	deploy.RegisterFlagCompletionFunc("platform-docker-tag", common.NoCompletions)
	deploy.RegisterFlagCompletionFunc("platform-chart-path", common.NoCompletions)
	deploy.RegisterFlagCompletionFunc("operator-docker-tag", common.NoCompletions)
	deploy.RegisterFlagCompletionFunc("operator-chart-path", common.NoCompletions)

	clean := &cobra.Command{
		Use:   "clean",
		Short: "Deletes the rig kind-cluster",
		Args:  cobra.NoArgs,
		RunE:  base.Register(func(c Cmd) any { return c.clean }),
		Annotations: map[string]string{
			base.OmitUser:    "",
			base.OmitProject: "",
		},
		ValidArgsFunction: common.NoCompletions,
	}
	kind.AddCommand(clean)

	parent.AddCommand(kind)
}
