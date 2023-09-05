package cmd

import (
	"context"

	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/internal/build"
	"github.com/rigdev/rig/internal/config"
	"github.com/rigdev/rig/internal/core"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/uuid"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// Used for flags.
var (
	projectFlag    string
	configFileFlag string
	rootCmd        = &cobra.Command{
		Use:   "rig-admin",
		Short: "Admin tool for managing a Rig setup",
	}
)

func init() {
	rootCmd.AddCommand(build.VersionCommand())

	rootCmd.PersistentFlags().StringVar(&projectFlag, "project", "rig", "project to target")
	rootCmd.PersistentFlags().StringVarP(&configFileFlag, "config", "c", "", "config file to use")
}

type ProjectID uuid.UUID

var options []fx.Option

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func register(f interface{}) func(cmd *cobra.Command, args []string) error {
	options = append(options,
		fx.Provide(f),
	)

	return func(cmd *cobra.Command, args []string) error {
		cfg, err := config.New(configFileFlag)
		if err != nil {
			return err
		}

		f := fx.New(
			fx.Decorate(func(ctx context.Context, pID ProjectID) context.Context {
				return auth.WithProjectID(ctx, uuid.UUID(pID))
			}),
			core.GetModule(cfg),
			fx.Supply(cmd),
			fx.Supply(args),
			fx.Provide(func() context.Context { return context.Background() }),
			fx.Provide(func() (ProjectID, error) {
				var pID uuid.UUID
				if projectFlag == "rig" {
					pID = auth.RigProjectID
				} else {
					id, err := uuid.Parse(projectFlag)
					if err != nil {
						return "", err
					}

					pID = id
				}

				return ProjectID(pID), nil
			}),
			fx.Provide(rig.NewClient),
			fx.Invoke(f),
			fx.NopLogger,
		)

		if err := f.Start(context.Background()); err != nil {
			return err
		}
		if err := f.Stop(context.Background()); err != nil {
			return err
		}
		return f.Err()
	}
}
