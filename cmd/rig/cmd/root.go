package cmd

import (
	"context"

	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/rig/cmd/auth"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule/root"
	"github.com/rigdev/rig/cmd/rig/cmd/cluster"
	"github.com/rigdev/rig/cmd/rig/cmd/cmd_config"
	"github.com/rigdev/rig/cmd/rig/cmd/config"
	"github.com/rigdev/rig/cmd/rig/cmd/dev"
	"github.com/rigdev/rig/cmd/rig/cmd/group"
	"github.com/rigdev/rig/cmd/rig/cmd/project"
	"github.com/rigdev/rig/cmd/rig/cmd/service_account"
	"github.com/rigdev/rig/cmd/rig/cmd/user"
	"github.com/rigdev/rig/pkg/build"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type RootCmd struct {
	fx.In

	Ctx    context.Context
	Rig    rig.Client
	Cfg    *cmd_config.Config
	Logger *zap.Logger

	Dev            dev.Cmd
	Capsule        root.Cmd
	Auth           auth.Cmd
	User           user.Cmd
	ServiceAccount service_account.Cmd
	Group          group.Cmd
	Cluster        cluster.Cmd
	Config         config.Cmd
	Project        project.Cmd
}

func (r RootCmd) Execute() error {
	rootCmd := &cobra.Command{
		Use:               "rig",
		Short:             "CLI tool for managing your Rig projects",
		PersistentPreRunE: r.preRun,
	}

	// database.Setup(rootCmd)
	// storage.Setup(rootCmd)
	r.Dev.Setup(rootCmd)
	r.Capsule.Setup(rootCmd)
	r.Auth.Setup(rootCmd)
	r.User.Setup(rootCmd)
	r.ServiceAccount.Setup(rootCmd)
	r.Group.Setup(rootCmd)
	r.Cluster.Setup(rootCmd)
	r.Config.Setup(rootCmd)
	r.Project.Setup(rootCmd)
	rootCmd.AddCommand(build.VersionCommand())

	return rootCmd.Execute()
}

func (r RootCmd) preRun(cmd *cobra.Command, args []string) error {
	return base.CheckAuth(cmd, r.Rig, r.Cfg)
}
