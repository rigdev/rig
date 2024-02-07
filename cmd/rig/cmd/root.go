package cmd

import (
	"context"
	"time"

	"connectrpc.com/connect"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-api/api/v1/environment"
	project_api "github.com/rigdev/rig-go-api/api/v1/project"
	"github.com/rigdev/rig-go-api/api/v1/project/settings"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/auth"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	capsule_root "github.com/rigdev/rig/cmd/rig/cmd/capsule/root"
	"github.com/rigdev/rig/cmd/rig/cmd/cluster"
	"github.com/rigdev/rig/cmd/rig/cmd/cmdconfig"
	"github.com/rigdev/rig/cmd/rig/cmd/config"
	"github.com/rigdev/rig/cmd/rig/cmd/dev"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/rigdev/rig/cmd/rig/cmd/group"
	"github.com/rigdev/rig/cmd/rig/cmd/project"
	"github.com/rigdev/rig/cmd/rig/cmd/serviceaccount"
	"github.com/rigdev/rig/cmd/rig/cmd/user"
	auth_service "github.com/rigdev/rig/cmd/rig/services/auth"
	"github.com/rigdev/rig/pkg/build"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Cmd struct {
	fx.In

	Rig    rig.Client
	Cfg    *cmdconfig.Config
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
	}
	rootCmd.PersistentFlags().VarP(&flags.Flags.OutputType, "output", "o", "output type. One of json,yaml,pretty.")
	rootCmd.PersistentFlags().StringVarP(&flags.Flags.Environment,
		"environment", "E", flags.Flags.Environment, "Override environment from the context")
	rootCmd.PersistentFlags().StringVarP(&flags.Flags.Project,
		"project", "P", flags.Flags.Project, "Override project from the context")
	rootCmd.PersistentFlags().BoolVarP(&flags.Flags.BasicAuth,
		"basic-auth", "B", flags.Flags.BasicAuth,
		"Use basic auth with environment in context. Otherwise use access token from the context")
	rootCmd.PersistentFlags().StringVarP(&flags.Flags.Host,
		"host", "H", flags.Flags.Host, "Override the rig host from the context")

	license := &cobra.Command{
		Use:               "license",
		Short:             "Get License Information for the current project",
		Args:              cobra.NoArgs,
		PersistentPreRunE: base.MakeInvokePreRunE(initCmd),
		RunE:              base.CtxWrap(cmd.getLicenseInfo),
		Annotations: map[string]string{
			auth_service.OmitProject:     "",
			auth_service.OmitEnvironment: "",
		},
	}
	rootCmd.AddCommand(license)

	version := &cobra.Command{
		Use:               "version",
		Short:             "print version information",
		PersistentPreRunE: base.MakeInvokePreRunE(initCmd),
		RunE:              base.CtxWrap(cmd.version),
		Annotations: map[string]string{
			auth_service.OmitProject:     "",
			auth_service.OmitEnvironment: "",
		},
	}
	version.Flags().BoolP("full", "v", false, "print full version")
	rootCmd.AddCommand(version)

	dev.Setup(rootCmd)
	capsule_root.Setup(rootCmd)
	auth.Setup(rootCmd)
	user.Setup(rootCmd)
	serviceaccount.Setup(rootCmd)
	group.Setup(rootCmd)
	cluster.Setup(rootCmd)
	config.Setup(rootCmd)
	project.Setup(rootCmd)

	cobra.EnableTraverseRunHooks = true
	return rootCmd.Execute()
}

func (c *Cmd) getLicenseInfo(ctx context.Context, cmd *cobra.Command, _ []string) error {
	var plan project_api.Plan
	var expiresAt *timestamppb.Timestamp

	resp, err := c.Rig.ProjectSettings().GetLicenseInfo(ctx, &connect.Request[settings.GetLicenseInfoRequest]{})
	if err != nil {
		cmd.Println("Unable to get license info", err)
		plan = project_api.Plan_PLAN_FREE
	} else {
		plan = resp.Msg.GetPlan()
		expiresAt = resp.Msg.GetExpiresAt()
	}

	if flags.Flags.OutputType != common.OutputTypePretty {
		obj := struct {
			Plan      string    `json:"plan" yaml:"plan"`
			ExpiresAt time.Time `json:"expires_at" yaml:"expires_at"`
		}{
			Plan:      plan.String(),
			ExpiresAt: expiresAt.AsTime(),
		}
		return common.FormatPrint(obj, flags.Flags.OutputType)
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

func (c *Cmd) version(ctx context.Context, cmd *cobra.Command, _ []string) error {
	full, err := cmd.Flags().GetBool("full")
	if err != nil {
		return err
	}

	if full {
		cmd.Println(build.VersionStringFull())
	} else {
		cmd.Println(build.VersionString())
	}

	if full {
		resp, err := c.Rig.Environment().List(ctx, &connect.Request[environment.ListRequest]{})
		if err != nil {
			cmd.Println("Unable to get platform version", err)
		} else {
			cmd.Println("Platform version:", resp.Msg.GetPlatformVersion())
		}
	}

	return nil
}
