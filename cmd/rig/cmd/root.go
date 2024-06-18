package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"connectrpc.com/connect"
	"github.com/jedib0t/go-pretty/v6/table"
	settings_api "github.com/rigdev/rig-go-api/api/v1/settings"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/auth"
	capsule_root "github.com/rigdev/rig/cmd/rig/cmd/capsule/root"
	"github.com/rigdev/rig/cmd/rig/cmd/cluster"
	"github.com/rigdev/rig/cmd/rig/cmd/completions"
	"github.com/rigdev/rig/cmd/rig/cmd/config"
	"github.com/rigdev/rig/cmd/rig/cmd/dev"
	"github.com/rigdev/rig/cmd/rig/cmd/environment"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/rigdev/rig/cmd/rig/cmd/git"
	"github.com/rigdev/rig/cmd/rig/cmd/group"
	"github.com/rigdev/rig/cmd/rig/cmd/noop"
	"github.com/rigdev/rig/cmd/rig/cmd/project"
	"github.com/rigdev/rig/cmd/rig/cmd/role"
	"github.com/rigdev/rig/cmd/rig/cmd/serviceaccount"
	"github.com/rigdev/rig/cmd/rig/cmd/settings"
	"github.com/rigdev/rig/cmd/rig/cmd/user"
	auth_service "github.com/rigdev/rig/cmd/rig/services/auth"
	"github.com/rigdev/rig/pkg/cli"
	"github.com/rigdev/rig/pkg/cli/scope"
	"github.com/rigdev/rig/pkg/cli/version"
	"github.com/rigdev/rig/pkg/license"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Cmd struct {
	fx.In

	Rig    rig.Client
	Scope  scope.Scope
	Logger *zap.Logger
}

var cmd Cmd

func initCmd(c Cmd) {
	cmd.Rig = c.Rig
	cmd.Scope = c.Scope
	cmd.Logger = c.Logger
}

func Run(s *cli.SetupContext) error {
	flags.InitFlags()

	rootCmd := &cobra.Command{
		Use:           "rig",
		Short:         "CLI tool for managing Rig",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	rootCmd.PersistentFlags().VarP(&flags.Flags.OutputType, "output", "o", "Output type. One of json,yaml,pretty.")
	rootCmd.PersistentFlags().StringVarP(&flags.Flags.Environment,
		"environment", "E", flags.Flags.Environment,
		"Select which environment to use. Can also be set with environment variable `RIG_ENVIRONMENT`")
	rootCmd.PersistentFlags().StringVarP(&flags.Flags.Project,
		"project", "P", flags.Flags.Project,
		"Select which project to use. Can also be set with environment variable `RIG_PROJECT`")
	rootCmd.PersistentFlags().StringVarP(&flags.Flags.Host,
		"host", "H", flags.Flags.Host,
		"Select which host to access the Rig Platform at. Should be of the form `http[s]://hostname:port/`."+
			" Can also be set with environment variable `RIG_HOST`")
	rootCmd.PersistentFlags().StringVarP(&flags.Flags.Context,
		"context", "C", flags.Flags.Context,
		"Select a context to use instead of the one currently set in the config.")

	if err := rootCmd.RegisterFlagCompletionFunc("project",
		cli.HackCtxWrapCompletion(cmd.completeProject, s)); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := rootCmd.RegisterFlagCompletionFunc("environment",
		cli.HackCtxWrapCompletion(cmd.completeEnvironment, s)); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := rootCmd.RegisterFlagCompletionFunc("context",
		cli.HackWrapCompletion(cmd.completeContext, s)); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := rootCmd.RegisterFlagCompletionFunc("output",
		completions.OutputType); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	rootCmd.AddGroup(
		&cobra.Group{
			ID:    common.CapsuleGroupID,
			Title: common.CapsuleGroupTitle,
		},
		&cobra.Group{
			ID:    common.ManagementGroupID,
			Title: common.ManagementGroupTitle,
		},
		&cobra.Group{
			ID:    common.AuthGroupID,
			Title: common.AuthGroupTitle,
		},
		&cobra.Group{
			ID:    common.OtherGroupID,
			Title: common.OtherGroupTitle,
		},
	)
	rootCmd.SetHelpCommandGroupID(common.OtherGroupID)
	rootCmd.SetCompletionCommandGroupID(common.OtherGroupID)

	license := &cobra.Command{
		Use:               "license",
		Short:             "Get license information",
		Args:              cobra.NoArgs,
		PersistentPreRunE: s.MakeInvokePreRunE(initCmd),
		RunE:              cli.CtxWrap(cmd.getLicenseInfo),
		Annotations: map[string]string{
			auth_service.OmitProject:     "",
			auth_service.OmitEnvironment: "",
		},
		GroupID: common.AuthGroupID,
	}
	rootCmd.AddCommand(license)

	version.Setup(rootCmd, s, common.OtherGroupID)
	dev.Setup(rootCmd, s)
	capsule_root.Setup(rootCmd, s)
	auth.Setup(rootCmd, s)
	user.Setup(rootCmd, s)
	serviceaccount.Setup(rootCmd, s)
	group.Setup(rootCmd, s)
	cluster.Setup(rootCmd, s)
	config.Setup(rootCmd, s)
	project.Setup(rootCmd, s)
	environment.Setup(rootCmd, s)
	settings.Setup(rootCmd, s)
	role.Setup(rootCmd, s)
	git.Setup(rootCmd, s)

	if s.AddTestCommand {
		noop.Setup(rootCmd, s)
	}

	cobra.EnableTraverseRunHooks = true

	if len(s.Args) > 0 {
		rootCmd.SetArgs(s.Args)
	}
	return rootCmd.Execute()
}

func (c *Cmd) getLicenseInfo(ctx context.Context, cmd *cobra.Command, _ []string) error {
	var expiresAt *timestamppb.Timestamp
	var userLimit license.UserLimit

	resp, err := c.Rig.Settings().GetLicenseInfo(ctx, &connect.Request[settings_api.GetLicenseInfoRequest]{})
	if err != nil {
		cmd.Println("Unable to get license info", err)
		userLimit = license.UnspecifiedNumUsers
	} else {
		fmt.Println("Got license info", resp.Msg.GetExpiresAt(), resp.Msg.GetUserLimit())
		expiresAt = resp.Msg.GetExpiresAt()
		userLimit = license.UserLimit(resp.Msg.GetUserLimit())
	}

	if flags.Flags.OutputType != common.OutputTypePretty {
		obj := struct {
			UserLimit license.UserLimit `json:"user_limit" yaml:"plan"`
			ExpiresAt time.Time         `json:"expires_at" yaml:"expires_at"`
		}{
			UserLimit: userLimit,
			ExpiresAt: expiresAt.AsTime(),
		}
		return common.FormatPrint(obj, flags.Flags.OutputType)
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{"Attribute", "Value"})
	t.AppendRows([]table.Row{
		{"User Limit", userLimit},
		{"Expires At", expiresAt.AsTime().Format("2006-01-02 15:04:05")},
	})

	cmd.Println(t.Render())

	return nil
}

func (c *Cmd) completeProject(
	ctx context.Context,
	cmd *cobra.Command,
	args []string,
	toComplete string,
	s *cli.SetupContext,
) ([]string, cobra.ShellCompDirective) {
	if err := s.ExecuteInvokes(cmd, args, initCmd); err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	return completions.Projects(ctx, c.Rig, toComplete)
}

func (c *Cmd) completeEnvironment(
	ctx context.Context,
	cmd *cobra.Command,
	args []string,
	toComplete string,
	s *cli.SetupContext,
) ([]string, cobra.ShellCompDirective) {
	if err := s.ExecuteInvokes(cmd, args, initCmd); err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	return completions.Environments(ctx, c.Rig, toComplete, flags.GetProject(c.Scope))
}

func (c *Cmd) completeContext(
	cmd *cobra.Command,
	args []string,
	toComplete string,
	s *cli.SetupContext,
) ([]string, cobra.ShellCompDirective) {
	if err := s.ExecuteInvokes(cmd, args, initCmd); err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	return completions.Contexts(toComplete, c.Scope.GetCfg())
}
