package environment

import (
	"context"
	"fmt"
	"os"

	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/completions"
	"github.com/rigdev/rig/cmd/rig/services/auth"
	"github.com/rigdev/rig/pkg/cli"
	"github.com/rigdev/rig/pkg/cli/scope"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var (
	useEnvironment    bool
	namespaceTemplate string
	force             bool
	failIfExists      bool
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
	environment := &cobra.Command{
		Use:               "environment",
		Aliases:           []string{"env"},
		Short:             "Manage Rig environments",
		PersistentPreRunE: s.MakeInvokePreRunE(initCmd),
		Annotations: map[string]string{
			auth.OmitEnvironment: "",
			auth.OmitProject:     "",
		},
		GroupID: common.ManagementGroupID,
	}

	listEnvironments := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all environments",
		Args:    cobra.NoArgs,
		RunE:    cli.CtxWrap(cmd.list),
	}
	environment.AddCommand(listEnvironments)

	createEnvironment := &cobra.Command{
		Use:   "create [environment] [cluster]",
		Short: "Create a new environment",
		Args:  cobra.MaximumNArgs(2),
		RunE:  cli.CtxWrap(cmd.create),
	}
	createEnvironment.Flags().BoolVar(&failIfExists, "fail-if-exists", false,
		"Fail the request if the environment already exists")
	createEnvironment.Flags().StringVar(&namespaceTemplate, "namespace-template", "",
		"Set the namespace-template used to generate namespaces for the given environment. ")
	createEnvironment.Flags().BoolVar(&useEnvironment, "use", false, "Use the created environment")
	if err := createEnvironment.RegisterFlagCompletionFunc("use", common.BoolCompletions); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	environment.AddCommand(createEnvironment)

	deleteEnvironment := &cobra.Command{
		Use:   "delete environment",
		Short: "Delete an environment",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: common.Complete(
			cli.HackCtxWrapCompletion(cmd.completeEnvironment, s),
			common.MaxArgsCompletionFilter(1)),
		RunE: cli.CtxWrap(cmd.delete),
	}
	deleteEnvironment.Flags().BoolVarP(&force, "force", "f", false,
		"Force deletion of all running capsules in the environment")
	environment.AddCommand(deleteEnvironment)

	parent.AddCommand(environment)
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

	return completions.Environments(ctx, c.Rig, toComplete)
}
