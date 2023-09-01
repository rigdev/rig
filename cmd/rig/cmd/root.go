package cmd

import (
	"github.com/rigdev/rig/cmd/rig/cmd/auth"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/config"
	"github.com/rigdev/rig/cmd/rig/cmd/database"
	"github.com/rigdev/rig/cmd/rig/cmd/group"
	"github.com/rigdev/rig/cmd/rig/cmd/project"
	"github.com/rigdev/rig/cmd/rig/cmd/service_account"
	"github.com/rigdev/rig/cmd/rig/cmd/storage"
	"github.com/rigdev/rig/cmd/rig/cmd/user"
	"github.com/rigdev/rig/internal/build"
	"github.com/spf13/cobra"
)

// Used for flags.
var (
	rootCmd = &cobra.Command{
		Use:   "rig",
		Short: "CLI tool for managing your Rig projects",
	}
)

// Execute executes the root command.
func Execute() error {
	auth.Setup(rootCmd)
	storage.Setup(rootCmd)
	capsule.Setup(rootCmd)
	database.Setup(rootCmd)
	service_account.Setup(rootCmd)
	user.Setup(rootCmd)
	group.Setup(rootCmd)
	project.Setup(rootCmd)
	config.Setup(rootCmd)
	rootCmd.AddCommand(build.VersionCommand())
	return rootCmd.Execute()
}
