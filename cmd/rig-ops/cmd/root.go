package cmd

import (
	"path/filepath"

	"github.com/rigdev/rig/cmd/rig-ops/cmd/base"
	"github.com/rigdev/rig/cmd/rig-ops/cmd/migrate"
	"github.com/rigdev/rig/cmd/rig-ops/cmd/plugins"
	"github.com/rigdev/rig/pkg/cli"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"k8s.io/client-go/util/homedir"
)

func Run() error {
	cli.AddOptions(
		fx.Provide(base.NewKubernetesClient),
		fx.Provide(base.NewKubernetesReader),
		fx.Provide(base.NewOperatorClient),
	)

	rootCmd := &cobra.Command{
		Use:           "rig-ops",
		Short:         "CLI tool for managing your Rig Clusters",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	rootCmd.PersistentFlags().StringVar(&base.Flags.KubeConfig,
		"kube-config", filepath.Join(homedir.HomeDir(), ".kube", "config"), "Path to your kubeconfig file")
	rootCmd.PersistentFlags().StringVar(&base.Flags.KubeContext,
		"kube-context", "", "The context to use from your kubeconfig file. Default is the current context")
	rootCmd.PersistentFlags().StringVarP(&base.Flags.KubeFile,
		"kube-file", "f", "", "A file of resources to use instead of k8s, for reading resources")
	rootCmd.PersistentFlags().StringVar(&base.Flags.RigContext,
		"rig-context", "", "The context to use from your rigconfig file. Default is the current context")
	rootCmd.PersistentFlags().StringVar(&base.Flags.RigConfig, "rig-config", "", "Path to your rigconfig file")
	rootCmd.PersistentFlags().StringVar(&base.Flags.OperatorConfig, "operator-config", "",
		"If given, will read the config file at the path and use that as an operator config. "+
			"If empty, will use the operator config of the running operator.")
	rootCmd.PersistentFlags().VarP(&base.Flags.OutputType, "output", "o", "output type. One of json,yaml,pretty.")

	migrate.Setup(rootCmd)
	plugins.Setup(rootCmd)

	return rootCmd.Execute()
}
