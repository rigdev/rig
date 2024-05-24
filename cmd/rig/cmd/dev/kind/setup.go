package kind

import (
	"github.com/docker/docker/client"
	"github.com/rigdev/rig/pkg/cli"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var (
	platformDockerTag string
	platformChartPath string
	platformValues    string
	operatorDockerTag string
	operatorChartPath string
	operatorValues    string
	prometheus        bool
	vpa               bool
	skipInit          bool
	installationID    string
)

type Cmd struct {
	fx.In

	DockerClient *client.Client
}

var cmd Cmd

func initCmd(c Cmd) {
	cmd.DockerClient = c.DockerClient
}

func Setup(parent *cobra.Command, s *cli.SetupContext) {
	kind := &cobra.Command{
		Use: "kind",
		Short: "The kind command is used to setup and manage a development kubernetes cluster running " +
			"Rig using Kind",
		PersistentPreRunE: s.MakeInvokePreRunE(initCmd),
	}

	create := &cobra.Command{
		Use:   "create",
		Short: "Create a rig cluster in Kind for local development",
		Args:  cobra.NoArgs,
		RunE:  cli.CtxWrap(cmd.create),
	}
	create.Flags().StringVarP(
		&installationID,
		"installation-id",
		"i", "",
		"Installation ID from the Rig Hub. "+
			"If set, the rig-platform will activate the hub installation, and inject the license",
	)
	create.Flags().StringVarP(
		&platformDockerTag,
		"platform-docker-tag", "p", "", "The rig-platform docker image tag. Defaults to latest.",
	)
	create.Flags().StringVar(
		&platformChartPath,
		"platform-chart-path", "", "If set, uses the helm chart at platform-chart-path to build rig-platform.",
	)
	create.Flags().StringVar(
		&platformValues,
		"platform-values", "", "If set, a custom values file for when installing the rig-platform chart.",
	)
	create.Flags().StringVarP(
		&operatorDockerTag,
		"operator-docker-tag", "r", "", "The rig-operator docker image tag. Defaults to latest.",
	)
	create.Flags().StringVar(
		&operatorChartPath,
		"operator-chart-path", "", "If set, uses the helm chart at operator-chart-path to build rig-operator.",
	)
	create.Flags().StringVar(
		&operatorValues,
		"operator-values", "", "If set, a custom values file for when installing the rig-operator chart.",
	)
	create.Flags().BoolVar(
		&prometheus,
		"prometheus", false,
		`If set, will install a Prometheus instance and Prometheus adapter using the Prometheus operator.
See https://github.com/prometheus-operator/prometheus-operator for information on the operator and its CRDs`,
	)
	create.Flags().BoolVar(
		&vpa,
		"vpa", false,
		`If set, will install the VerticalPodAutoscaler CRD + Recommender in the cluster.
The operator will be configured to spawn a VerticalPodAutoscaler resource per capsule.`,
	)
	create.Flags().BoolVar(
		&skipInit, "skip-init", false, "If set, will skip prompting for rig initialization after platform is up and running.",
	)
	kind.AddCommand(create)

	deploy := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy a new (or specific) version of Rig to the kind cluster",
		Args:  cobra.NoArgs,
		RunE:  cli.CtxWrap(cmd.deploy),
	}
	kind.AddCommand(deploy)
	deploy.Flags().StringVarP(
		&platformDockerTag,
		"platform-docker-tag", "p", "", "The rig-platform docker image tag. Defaults to latest.",
	)
	deploy.Flags().StringVar(
		&platformChartPath,
		"platform-chart-path", "", "If set, uses the helm chart at platform-chart-path to build rig-platform.",
	)
	deploy.Flags().StringVar(
		&platformValues,
		"platform-values", "", "If set, a custom values file for when installing the rig-platform chart.",
	)
	deploy.Flags().StringVarP(
		&operatorDockerTag,
		"operator-docker-tag", "r", "", "The rig-operator docker image tag. Defaults to latest.",
	)
	deploy.Flags().StringVar(
		&operatorChartPath,
		"operator-chart-path", "", "If set, uses the helm chart at operator-chart-path to build rig-operator.",
	)
	deploy.Flags().StringVar(
		&operatorValues,
		"operator-values", "", "If set, a custom values file for when installing the rig-operator chart.",
	)
	deploy.Flags().BoolVar(
		&prometheus,
		"prometheus", false,
		//nolint:lll
		"If set, will instruct the operator to create Prometheus ServiceMonitors for each capsule, scraping ports named 'metrics'.",
	)
	deploy.Flags().BoolVar(
		&vpa,
		"vpa", false,
		`If set, the operator will be configured to spawn a VerticalPodAutoscaler resource per capsule.`,
	)

	clean := &cobra.Command{
		Use:   "clean",
		Short: "Deletes the rig kind-cluster",
		Args:  cobra.NoArgs,
		RunE:  cli.CtxWrap(cmd.clean),
	}
	kind.AddCommand(clean)

	runInit := &cobra.Command{
		Use:   "run-init",
		Short: "Runs the 'init' command on the rig-platform which creates the first admin user and project",
		Args:  cobra.NoArgs,
		RunE:  cmd.runInit,
	}
	kind.AddCommand(runInit)

	parent.AddCommand(kind)
}
