package kind

import (
	"github.com/docker/docker/client"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
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
}

var cmd Cmd

func initCmd(c Cmd) {
	cmd.DockerClient = c.DockerClient
}

func Setup(parent *cobra.Command) {
	kind := &cobra.Command{
		Use:               "kind",
		Short:             "The kind command is used to setup and manage a development kubernetes cluster running Rig using Kind",
		PersistentPreRunE: base.MakeInvokePreRunE(initCmd),
	}

	create := &cobra.Command{
		Use:   "create",
		Short: "Create a rig cluster in Kind for local development",
		Args:  cobra.NoArgs,
		RunE:  base.CtxWrap(cmd.create),
		Annotations: map[string]string{
			base.OmitUser:    "",
			base.OmitProject: "",
		},
	}
	create.Flags().StringVarP(&platformDockerTag, "platform-docker-tag", "p", "", "The rig-platform docker image tag. Defaults to latest.")
	create.Flags().StringVar(&platformChartPath, "platform-chart-path", "", "If set, uses the helm chart at platform-chart-path to build rig-platform.")
	create.Flags().StringVarP(&operatorDockerTag, "operator-docker-tag", "r", "", "The rig-operator docker image tag. Defaults to latest.")
	create.Flags().StringVar(&operatorChartPath, "operator-chart-path", "", "If set, uses the helm chart at operator-chart-path to build rig-operator.")
	kind.AddCommand(create)

	deploy := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy a new (or specific) version of Rig to the kind cluster",
		Args:  cobra.NoArgs,
		RunE:  base.CtxWrap(cmd.deploy),
		Annotations: map[string]string{
			base.OmitUser:    "",
			base.OmitProject: "",
		},
	}
	kind.AddCommand(deploy)
	deploy.Flags().StringVarP(&platformDockerTag, "platform-docker-tag", "p", "", "The rig-platform docker image tag. Defaults to latest.")
	deploy.Flags().StringVar(&platformChartPath, "platform-chart-path", "", "If set, uses the helm chart at platform-chart-path to build rig-platform.")
	deploy.Flags().StringVarP(&operatorDockerTag, "operator-docker-tag", "r", "", "The rig-operator docker image tag. Defaults to latest.")
	deploy.Flags().StringVar(&operatorChartPath, "operator-chart-path", "", "If set, uses the helm chart at operator-chart-path to build rig-operator.")

	clean := &cobra.Command{
		Use:   "clean",
		Short: "Deletes the rig kind-cluster",
		Args:  cobra.NoArgs,
		RunE:  base.CtxWrap(cmd.clean),
		Annotations: map[string]string{
			base.OmitUser:    "",
			base.OmitProject: "",
		},
	}
	kind.AddCommand(clean)

	parent.AddCommand(kind)
}
