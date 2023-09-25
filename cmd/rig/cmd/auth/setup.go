package auth

import (
	"context"

	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/rigdev/rig/cmd/rig/cmd/cmd_config"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var (
	authPassword       string
	authUserIdentifier string
	redirectAddr       string
)

var outputJSON bool

type Cmd struct {
	fx.In

	Ctx context.Context
	Rig rig.Client
	Cfg *cmd_config.Config
}

func (c Cmd) Setup(parent *cobra.Command) {
	auth := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication for the current user",
	}

	login := &cobra.Command{
		Use:   "login",
		Short: "Login with user identifier and password",
		Args:  cobra.NoArgs,
		Annotations: map[string]string{
			base.OmitUser:    "",
			base.OmitProject: "",
		},
		RunE: c.login,
	}
	login.Flags().StringVarP(&authUserIdentifier, "user", "u", "", "useridentifier [username | email | phone number]")
	login.Flags().StringVarP(&authPassword, "password", "p", "", "password of the user")
	auth.AddCommand(login)

	get := &cobra.Command{
		Use:   "get",
		Short: "Get user information associated with the current user",
		Args:  cobra.NoArgs,
		RunE:  c.get,
		Annotations: map[string]string{
			base.OmitProject: "",
		},
	}
	get.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	auth.AddCommand(get)

	getAuthConfig := &cobra.Command{
		Use:   "get-auth-config {project-id | project-name}",
		Short: "Get the authorization config with allowed login methods and configurations",
		Args:  cobra.MaximumNArgs(1),
		RunE:  c.getAuthConfig,
		Annotations: map[string]string{
			base.OmitProject: "",
		},
	}
	getAuthConfig.Flags().StringVarP(&redirectAddr, "redirect-addr", "r", "", "redirect address for oauth2")
	getAuthConfig.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	auth.AddCommand(getAuthConfig)

	parent.AddCommand(auth)
}
