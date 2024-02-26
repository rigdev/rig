package cmd

import (
	"path/filepath"

	"github.com/rigdev/rig/cmd/rig-ops/cmd/base"
	"github.com/rigdev/rig/cmd/rig-ops/cmd/migrate"
	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"
)

func Run() error {
	rootCmd := &cobra.Command{
		Use:   "rig-ops",
		Short: "CLI tool for managing your Rig Clusters",
	}
	rootCmd.PersistentFlags().StringVar(&base.Flags.KubeConfig,
		"kube-config", filepath.Join(homedir.HomeDir(), ".kube", "config"), "Path to your kubeconfig file")
	rootCmd.PersistentFlags().StringVar(&base.Flags.KubeContext,
		"kube-context", "", "The context to use from your kubeconfig file. Default is the current context")
	rootCmd.PersistentFlags().StringVarP(&base.Flags.Namespace, "namespace", "n", "", "The k8s namespace to migrate from")
	rootCmd.PersistentFlags().StringVarP(&base.Flags.KubeFile,
		"kube-file", "f", "", "A file of resources to use instead of k8s, for reading resources")
	rootCmd.PersistentFlags().StringVar(&base.Flags.RigContext,
		"rig-context", "", "The context to use from your rigconfig file. Default is the current context")
	rootCmd.PersistentFlags().StringVar(&base.Flags.RigConfig, "rig-config", "", "Path to your rigconfig file")
	rootCmd.PersistentFlags().StringVarP(&base.Flags.Project, "project", "p", "", "The project to migrate to")

	migrate.Setup(rootCmd)

	return rootCmd.Execute()
}
