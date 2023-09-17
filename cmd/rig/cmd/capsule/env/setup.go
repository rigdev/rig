package env

import (
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/spf13/cobra"
)

func Setup(parent *cobra.Command) *cobra.Command {
	env := &cobra.Command{
		Use:   "env",
		Short: "Manage environment variables for the capsule",
	}

	envSet := &cobra.Command{
		Use:               "set key value",
		Short:             "Set an environment variable",
		Args:              cobra.ExactArgs(2),
		RunE:              base.Register(set),
		ValidArgsFunction: common.NoCompletions,
	}
	env.AddCommand(envSet)

	envGet := &cobra.Command{
		Use:               "get [key]",
		Short:             "Get an environment variable",
		Args:              cobra.MaximumNArgs(1),
		RunE:              base.Register(get),
		ValidArgsFunction: common.Complete(capsule.EnvCompletions, common.MaxArgsCompletionFilter(1)),
	}
	env.AddCommand(envGet)

	envRemove := &cobra.Command{
		Use:               "remove [key]",
		Short:             "Remove an environment variable",
		Args:              cobra.ExactArgs(1),
		RunE:              base.Register(remove),
		ValidArgsFunction: common.Complete(capsule.EnvCompletions, common.MaxArgsCompletionFilter(1)),
	}
	env.AddCommand(envRemove)

	parent.AddCommand(env)

	return env
}
