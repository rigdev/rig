package main

import (
	"context"

	"github.com/rigdev/rig/internal/build"
	"github.com/rigdev/rig/pkg/config"
	"github.com/rigdev/rig/pkg/manager"
	"github.com/spf13/cobra"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

const (
	flagConfigFile = "config-file"
)

func main() {
	c := &cobra.Command{
		Use:   "rig-operator",
		Short: "operator for the rig.dev CRDs",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd, args)
		},
	}

	flags := c.PersistentFlags()
	flags.StringP(flagConfigFile, "c", "/etc/rig-operator/config.yaml", "path to rig-operator config file")

	c.AddCommand(build.VersionCommand())

	ctx := context.Background()
	c.ExecuteContext(ctx)
}

func run(cmd *cobra.Command, args []string) error {
	cfgFile, err := cmd.Flags().GetString(flagConfigFile)
	if err != nil {
		return err
	}

	scheme := manager.NewScheme()

	cfg, err := config.NewLoader(scheme).Load(cfgFile)
	if err != nil {
		return err
	}

	ctrl.SetLogger(zap.New(zap.UseDevMode(cfg.DevModeEnabled)))

	mgr, err := manager.NewManager(cfg, scheme)
	if err != nil {
		return err
	}

	return mgr.Start(cmd.Context())
}
