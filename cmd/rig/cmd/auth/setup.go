package auth

import (
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/services/auth"
	"github.com/rigdev/rig/pkg/cli"
	"github.com/rigdev/rig/pkg/cli/scope"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var (
	authPassword       string
	authUserIdentifier string
)

type Cmd struct {
	fx.In

	Rig      rig.Client
	Scope    scope.Scope
	Prompter common.Prompter
}

var cmd Cmd

func initCmd(c Cmd) {
	cmd.Rig = c.Rig
	cmd.Scope = c.Scope
	cmd.Prompter = c.Prompter
}

func Setup(parent *cobra.Command, s *cli.SetupContext) {
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Authenticate with users or service accounts",
		Annotations: map[string]string{
			auth.OmitProject:     "",
			auth.OmitEnvironment: "",
		},
		GroupID: common.AuthGroupID,
	}

	login := &cobra.Command{
		Use:   "login",
		Short: "Login with user identifier and password",
		Args:  cobra.NoArgs,
		Annotations: map[string]string{
			auth.OmitUser: "",
		},
		PersistentPreRunE: s.MakeInvokePreRunE(initCmd),
		RunE:              cli.CtxWrap(cmd.login),
	}
	login.Flags().StringVarP(&authUserIdentifier, "user", "u", "", "useridentifier [username | email | phone number]")
	login.Flags().StringVarP(&authPassword, "password", "p", "", "password of the user")
	authCmd.AddCommand(login)

	activateServiceAccount := &cobra.Command{
		Use:   "activate-service-account",
		Short: "Activate a service-account and store auth token in config file",
		Long: `
Activate a Service account by signing in using RIG_CLIENT_ID and RIG_CLIENT_SECRET.

The command expects there to be a host already configured. If not, --host/-H or RIG_HOST can be
used to provide a host for the command.

After activation, the account is able to refresh the token until the session expires or becomes
invalidated by the server.`,
		Args: cobra.NoArgs,
		Annotations: map[string]string{
			auth.OmitUser: "",
		},
		PersistentPreRunE: s.MakeInvokePreRunE(
			fx.Annotate(
				initServiceAccountContext,
				fx.As(new(cli.ContextDependency)),
				fx.ResultTags(`group:"context_dependencies"`),
			), initCmd),
		RunE: cli.CtxWrap(cmd.activateServiceAccount),
	}
	authCmd.AddCommand(activateServiceAccount)

	get := &cobra.Command{
		Use:               "get",
		Short:             "Get user information associated with the current user",
		Args:              cobra.NoArgs,
		PersistentPreRunE: s.MakeInvokePreRunE(initCmd),
		RunE:              cli.CtxWrap(cmd.get),
	}
	authCmd.AddCommand(get)

	parent.AddCommand(authCmd)
}
