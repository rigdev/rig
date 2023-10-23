package cmd

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/jedib0t/go-pretty/v6/table"
	project_api "github.com/rigdev/rig-go-api/api/v1/project"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
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
	"google.golang.org/protobuf/types/known/timestamppb"
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

	license := &cobra.Command{
		Use:               "license",
		Short:             "Get License Information for the current project",
		Args:              cobra.NoArgs,
		RunE:              r.getLicenseInfo,
		ValidArgsFunction: common.NoCompletions,
		Annotations: map[string]string{
			base.OmitProject: "",
		},
	}
	rootCmd.AddCommand(license)

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

func (c RootCmd) getLicenseInfo(cmd *cobra.Command, args []string) error {
	var plan project_api.Plan
	var expiresAt *timestamppb.Timestamp

	resp, err := c.Rig.Project().GetLicenseInfo(c.Ctx, &connect.Request[project_api.GetLicenseInfoRequest]{})
	if err != nil {
		cmd.Println("Unable to get license info", err)
		plan = project_api.Plan_PLAN_FREE
	} else {
		plan = resp.Msg.GetPlan()
		expiresAt = resp.Msg.GetExpiresAt()
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{"Attribute", "Value"})
	t.AppendRows([]table.Row{
		{"Plan", plan.String()},
		{"Expires At", expiresAt.AsTime().Format("2006-01-02 15:04:05")},
	})

	cmd.Println(t.Render())

	return nil
}
