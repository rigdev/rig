package cmd

import (
	"github.com/rigdev/rig/cmd/rig/cmd/auth"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule"
	capsuleBuild "github.com/rigdev/rig/cmd/rig/cmd/capsule/build"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule/env"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule/instance"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule/mount"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule/network"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule/resource"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule/rollout"
	"github.com/rigdev/rig/cmd/rig/cmd/cluster"
	"github.com/rigdev/rig/cmd/rig/cmd/config"
	"github.com/rigdev/rig/cmd/rig/cmd/dev"
	"github.com/rigdev/rig/cmd/rig/cmd/group"
	"github.com/rigdev/rig/cmd/rig/cmd/project"
	"github.com/rigdev/rig/cmd/rig/cmd/service_account"
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
	// database.Setup(rootCmd)
	// storage.Setup(rootCmd)
	auth.Setup(rootCmd)
	user.Setup(rootCmd)
	service_account.Setup(rootCmd)
	group.Setup(rootCmd)
	project.Setup(rootCmd)
	config.Setup(rootCmd)
	cluster.Setup(rootCmd)
	dev.Setup(rootCmd)

	capsuleCmd := capsule.Setup(rootCmd)
	resource.Setup(capsuleCmd)
	capsuleBuild.Setup(capsuleCmd)
	instance.Setup(capsuleCmd)
	network.Setup(capsuleCmd)
	rollout.Setup(capsuleCmd)
	env.Setup(capsuleCmd)
	mount.Setup(capsuleCmd)

	rootCmd.AddCommand(build.VersionCommand())
	return rootCmd.Execute()
}
