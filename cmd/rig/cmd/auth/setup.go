package auth

import (
	"os"

	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/spf13/cobra"
)

var (
	authUser     string
	authPassword string
	redirectAddr string
)

var outputJSON bool

func Setup(parent *cobra.Command) {
	auth := &cobra.Command{
		Use: "auth",
	}

	login := &cobra.Command{
		Use:   "login",
		Short: "Login with username and password",
		Args:  cobra.NoArgs,
		RunE:  base.Register(AuthLogin),
	}
	login.PersistentFlags().StringVarP(&authUser, "user", "u", os.Getenv("RIG_USER"), "name of the user, can be either email or username")
	login.PersistentFlags().StringVarP(&authPassword, "password", "p", os.Getenv("RIG_PASSWORD"), "password for the user")
	auth.AddCommand(login)

	get := &cobra.Command{
		Use:   "get",
		Short: "Get user information associated with the current user",
		Args:  cobra.NoArgs,
		RunE:  base.Register(AuthGet),
	}
	get.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	auth.AddCommand(get)

	getAuthConfig := &cobra.Command{
		Use:   "get-auth-config {project-id | project-name}",
		Short: "Get the authorization config with allowed login methods and configurations",
		Args:  cobra.MaximumNArgs(1),
		RunE:  base.Register(AuthGetAuthConfig),
	}
	getAuthConfig.Flags().StringVarP(&redirectAddr, "redirect-addr", "r", "", "redirect address for oauth2")
	getAuthConfig.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	auth.AddCommand(getAuthConfig)

	parent.AddCommand(auth)
}
