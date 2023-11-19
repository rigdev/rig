package auth

import (
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/rigdev/rig/cmd/rig/cmd/cmd_config"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var (
	authPassword       string
	authUserIdentifier string
)

type Cmd struct {
	fx.In

	Rig rig.Client
	Cfg *cmd_config.Config
}

var cmd Cmd

func initCmd(c Cmd) {
	cmd.Rig = c.Rig
	cmd.Cfg = c.Cfg
}

func Setup(parent *cobra.Command) {
	auth := &cobra.Command{
		Use:               "auth",
		Short:             "Manage authentication for the current user",
		PersistentPreRunE: base.MakeInvokePreRunE(initCmd),
	}

	login := &cobra.Command{
		Use:   "login",
		Short: "Login with user identifier and password",
		Args:  cobra.NoArgs,
		Annotations: map[string]string{
			base.OmitUser:    "",
			base.OmitProject: "",
		},
		ValidArgsFunction: common.NoCompletions,
		RunE:              base.CtxWrap(cmd.login),
	}
	login.Flags().StringVarP(&authUserIdentifier, "user", "u", "", "useridentifier [username | email | phone number]")
	login.Flags().StringVarP(&authPassword, "password", "p", "", "password of the user")
	login.RegisterFlagCompletionFunc("user", common.NoCompletions)
	login.RegisterFlagCompletionFunc("password", common.NoCompletions)
	auth.AddCommand(login)

	get := &cobra.Command{
		Use:   "get",
		Short: "Get user information associated with the current user",
		Args:  cobra.NoArgs,
		RunE:  base.CtxWrap(cmd.get),
		Annotations: map[string]string{
			base.OmitProject: "",
		},
		ValidArgsFunction: common.NoCompletions,
	}
	auth.AddCommand(get)

	parent.AddCommand(auth)
}
