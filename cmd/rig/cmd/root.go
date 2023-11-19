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
	capsule_root "github.com/rigdev/rig/cmd/rig/cmd/capsule/root"
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

type Cmd struct {
	fx.In

	Rig    rig.Client
	Cfg    *cmd_config.Config
	Logger *zap.Logger
}

var cmd Cmd

func initCmd(c Cmd) {
	cmd.Rig = c.Rig
	cmd.Cfg = c.Cfg
	cmd.Logger = c.Logger
}

func Run() error {
	rootCmd := &cobra.Command{
		Use:   "rig",
		Short: "CLI tool for managing your Rig projects",
		PersistentPreRunE: base.MakeInvokePreRunE(
			initCmd,
			func(ctx context.Context, c *cobra.Command, args []string) error {
				return cmd.preRun(ctx, c, args)
			},
		),
	}
	rootCmd.PersistentFlags().VarP(&base.Flags.OutputType, "output", "o", "output type. One of json,yaml,pretty.")

	license := &cobra.Command{
		Use:               "license",
		Short:             "Get License Information for the current project",
		Args:              cobra.NoArgs,
		RunE:              base.CtxWrap(cmd.getLicenseInfo),
		ValidArgsFunction: common.NoCompletions,
		Annotations: map[string]string{
			base.OmitProject: "",
		},
	}
	rootCmd.AddCommand(license)

	dev.Setup(rootCmd)
	capsule_root.Setup(rootCmd)
	auth.Setup(rootCmd)
	user.Setup(rootCmd)
	service_account.Setup(rootCmd)
	group.Setup(rootCmd)
	cluster.Setup(rootCmd)
	config.Setup(rootCmd)
	project.Setup(rootCmd)
	rootCmd.AddCommand(build.VersionCommand())

	cobra.EnableTraverseRunHooks = true
	return rootCmd.Execute()
}

func (c *Cmd) preRun(ctx context.Context, cmd *cobra.Command, args []string) error {
	return base.CheckAuth(ctx, cmd, c.Rig, c.Cfg)
}

func (c *Cmd) getLicenseInfo(ctx context.Context, cmd *cobra.Command, args []string) error {
	var plan project_api.Plan
	var expiresAt *timestamppb.Timestamp

	resp, err := c.Rig.Project().GetLicenseInfo(ctx, &connect.Request[project_api.GetLicenseInfoRequest]{})
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
