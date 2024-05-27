package environment

import (
	"context"
	"fmt"
	"os"
	"strings"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/cluster"
	"github.com/rigdev/rig-go-api/api/v1/environment"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/completions"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
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
	ephemeral         bool
	addProjects       []string
	removeProjects    []string
	projectFilter     string
	global            bool
	updateGlobal      bool
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
	listEnvironments.Flags().BoolVar(&ephemeral, "exclude-ephemeral", false, "Exclude ephemeral environments")
	if err := listEnvironments.RegisterFlagCompletionFunc("exclude-ephemeral", common.BoolCompletions); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	listEnvironments.Flags().StringVar(&projectFilter, "project-filter", "",
		"Only include environments in which the given project is active")
	if err := listEnvironments.RegisterFlagCompletionFunc("project-filter",
		cli.HackCtxWrapCompletion(cmd.completeProject, s)); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	environment.AddCommand(listEnvironments)

	createEnvironment := &cobra.Command{
		Use: "create [environment] [cluster]",
		ValidArgsFunction: common.ChainCompletions([]int{1, 2}, common.NilCompletions,
			cli.HackCtxWrapCompletion(cmd.completeCluster, s)),
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
	createEnvironment.Flags().BoolVar(&ephemeral, "ephemeral", false, "Create an ephemeral environment")
	if err := createEnvironment.RegisterFlagCompletionFunc("ephemeral", common.BoolCompletions); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	createEnvironment.Flags().BoolVar(&global, "global", false, "Create a global environment.")
	if err := createEnvironment.RegisterFlagCompletionFunc("global", common.BoolCompletions); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	createEnvironment.Flags().StringSliceVar(&addProjects, "projects", nil,
		"Set the active projects for the environment. "+
			"If left empty, all projects will be active.")
	if err := createEnvironment.RegisterFlagCompletionFunc("projects",
		cli.HackCtxWrapCompletion(cmd.completeProject, s)); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	createEnvironment.MarkFlagsMutuallyExclusive("global", "projects")
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

	update := &cobra.Command{
		Use: "update [environment]",
		ValidArgsFunction: common.Complete(cli.HackCtxWrapCompletion(cmd.completeEnvironment, s),
			common.MaxArgsCompletionFilter(1)),
		Short: "Update an environment with fields provided by the flags.",
		Args:  cobra.ExactArgs(1),
		RunE:  cli.CtxWrap(cmd.update),
	}

	update.Flags().StringSliceVar(&addProjects, "add-projects", nil,
		"Add one or more projects to the environment")
	if err := update.RegisterFlagCompletionFunc("add-projects",
		cli.HackCtxWrapCompletion(cmd.completeProject, s)); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	update.Flags().StringSliceVar(&removeProjects, "remove-projects", nil,
		"Remove one or more projects from the environment")
	if err := update.RegisterFlagCompletionFunc("remove-projects",
		cli.HackCtxWrapCompletion(cmd.completeRemoveProject, s)); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	update.Flags().BoolVar(&updateGlobal, "set-global", false, "Set the environment as global")
	if err := update.RegisterFlagCompletionFunc("set-global", common.BoolCompletions); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	environment.AddCommand(update)

	removeProject := &cobra.Command{
		Use: "remove-project [environment] [project]",
		ValidArgsFunction: common.ChainCompletions([]int{1, 2},
			cli.HackCtxWrapCompletion(cmd.completeEnvironment, s),
			cli.HackCtxWrapCompletion(cmd.completeRemoveProject, s)),
		Short: "Remove a project from an environment. If this is the last project removed from the environment, " +
			"the environment will be available to all others",
		Args: cobra.ExactArgs(2),
		RunE: cli.CtxWrap(cmd.removeProject),
	}
	environment.AddCommand(removeProject)

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

	return completions.Environments(ctx, c.Rig, toComplete, flags.GetProject(c.Scope))
}

func (c *Cmd) completeProject(ctx context.Context,
	cmd *cobra.Command,
	args []string,
	toComplete string,
	s *cli.SetupContext,
) ([]string, cobra.ShellCompDirective) {
	if err := s.ExecuteInvokes(cmd, args, initCmd); err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	projectArgs := strings.Split(toComplete, ",")
	if len(projectArgs) > 1 {
		toComplete = projectArgs[len(projectArgs)-1]
	}

	if strings.HasSuffix(toComplete, ",") {
		toComplete = ""
	}

	return completions.Projects(ctx, c.Rig, toComplete)
}

func (c *Cmd) completeRemoveProject(ctx context.Context,
	cmd *cobra.Command,
	args []string,
	toComplete string,
	s *cli.SetupContext,
) ([]string, cobra.ShellCompDirective) {
	if err := s.ExecuteInvokes(cmd, args, initCmd); err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	envID := args[0]
	envs, err := c.Rig.Environment().List(ctx, connect.NewRequest(&environment.ListRequest{}))
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var projects []string
	for _, env := range envs.Msg.GetEnvironments() {
		if env.GetEnvironmentId() == envID {
			projects = env.GetActiveProjects()
		}
	}

	var completions []string
	for _, project := range projects {
		if strings.HasPrefix(project, toComplete) {
			completions = append(completions, project)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

func (c *Cmd) completeCluster(ctx context.Context,
	cmd *cobra.Command,
	args []string,
	toComplete string,
	s *cli.SetupContext,
) ([]string, cobra.ShellCompDirective) {
	if err := s.ExecuteInvokes(cmd, args, initCmd); err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	clustersResp, err := c.Rig.Cluster().List(ctx, connect.NewRequest(&cluster.ListRequest{}))
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var completions []string
	for _, cl := range clustersResp.Msg.GetClusters() {
		if strings.HasPrefix(cl.GetClusterId(), toComplete) {
			completions = append(completions, cl.GetClusterId())
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}
