package main

import (
	"context"
	"fmt"

	"github.com/rigdev/rig/internal/build"
	"github.com/rigdev/rig/internal/config"
	"github.com/rigdev/rig/internal/core"
	"github.com/rigdev/rig/internal/handler"
	"github.com/rigdev/rig/internal/handler/registry"
	"github.com/rigdev/rig/internal/service/operator"
	pkg_service "github.com/rigdev/rig/pkg/service"
	"github.com/spf13/cobra"
	"go.uber.org/dig"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

const (
	flagRootConfig = "config"
)

var verbose = false

func createRootCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use: "rig-server",
		RunE: func(cmd *cobra.Command, args []string) error {
			configFile, err := cmd.Flags().GetString(flagRootConfig)
			if err != nil {
				return err
			}

			cfg, err := config.New(configFile)
			if err != nil {
				return err
			}

			var opts []fx.Option
			if cfg.Registry.Enabled {
				opts = append(opts, fx.Invoke(func(_ *registry.Server) {}))
			}

			if cfg.Cluster.Type == config.ClusterTypeKubernetes {
				opts = append(opts, fx.Invoke(func(_ operator.Service) {}))
			}

			f := fx.New(
				core.GetModule(cfg),
				handler.Module,
				fx.Invoke(
					func(cfg config.Config, s *pkg_service.Server) {
						s.EmbeddedFileServer()
						s.Init()
					},
				),
				fx.WithLogger(func(logger *zap.Logger) fxevent.Logger {
					if !verbose {
						logger = logger.WithOptions(zap.IncreaseLevel(zap.PanicLevel))
					}
					return &fxevent.ZapLogger{Logger: logger}
				}),
				fx.Options(opts...),
			)

			run := func() error {
				startCtx, cancel := context.WithTimeout(context.Background(), f.StartTimeout())
				defer cancel()

				if err := f.Start(startCtx); err != nil {
					return err
				}

				sig := <-f.Wait()
				var err error
				if sig.ExitCode != 0 {
					err = fmt.Errorf("aborted: signal %v", sig.Signal.String())
				}

				stopCtx, cancel := context.WithTimeout(context.Background(), f.StopTimeout())
				defer cancel()

				if err := f.Stop(stopCtx); err != nil {
					return err
				}

				return err
			}

			return dig.RootCause(run())
		},
	}

	pflags := cmd.PersistentFlags()
	pflags.StringP("config", "c", "", "config file path")
	pflags.BoolVarP(&verbose, "verbose", "v", false, "enable verbose error logging")

	cmd.AddCommand(build.VersionCommand())

	return cmd
}
