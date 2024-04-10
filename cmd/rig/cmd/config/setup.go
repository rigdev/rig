package config

import (
	"context"
	"fmt"
	"os"
	"strings"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/environment"
	"github.com/rigdev/rig-go-api/api/v1/project"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/cmdconfig"
	"github.com/rigdev/rig/cmd/rig/services/auth"
	"github.com/rigdev/rig/pkg/cli"
	"github.com/rigdev/rig/pkg/cli/scope"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

type CmdWScope struct {
	fx.In

	Rig        rig.Client
	PromptInfo *cli.PromptInformation
	Scope      scope.Scope
	Prompter   common.Prompter
}

var minify bool

var (
	field       string
	value       string
	contextName string
)

var cmdWScope CmdWScope

func initCmdWScope(c CmdWScope) {
	cmdWScope = c
}

type CmdNoScope struct {
	fx.In

	PromptInfo  *cli.PromptInformation
	Cfg         *cmdconfig.Config
	Prompter    common.Prompter
	Interactive scope.Interactive
}

var cmdNoScope CmdNoScope

func initCmdNoScope(c CmdNoScope) {
	cmdNoScope = c
}

func Setup(parent *cobra.Command, s *cli.SetupContext) {
	config := &cobra.Command{
		Use:   "config",
		Short: "Manage Rig CLI configuration",
		Annotations: map[string]string{
			auth.OmitProject:     "",
			auth.OmitEnvironment: "",
		},
	}

	init := &cobra.Command{
		Use:               "init",
		Short:             "Initialize a new context",
		Args:              cobra.NoArgs,
		RunE:              cmdNoScope.init,
		PersistentPreRunE: s.MakeInvokePreRunE(initCmdNoScope),
		Annotations: map[string]string{
			auth.OmitUser: "",
		},
	}
	init.Flags().StringVar(&contextName, "name", "", "name of the context to create")
	config.AddCommand(init)

	deleteContext := &cobra.Command{
		Use:               "delete [context]",
		Short:             "Delete a context",
		Args:              cobra.ExactArgs(1),
		RunE:              cmdNoScope.delete,
		PersistentPreRunE: s.MakeInvokePreRunE(initCmdNoScope),
		ValidArgsFunction: common.Complete(
			cli.HackWrapCompletion(cmdNoScope.completions, s),
			common.MaxArgsCompletionFilter(1),
		),
		Annotations: map[string]string{
			auth.OmitUser: "",
		},
	}
	config.AddCommand(deleteContext)

	useContext := &cobra.Command{
		Use:               "use-context [context]",
		Short:             "Change the current context to use",
		Args:              cobra.MaximumNArgs(1),
		RunE:              cmdNoScope.useContext,
		PersistentPreRunE: s.MakeInvokePreRunE(initCmdNoScope),
		Annotations: map[string]string{
			auth.OmitUser: "",
		},
		ValidArgsFunction: common.Complete(
			cli.HackWrapCompletion(cmdNoScope.completions, s),
			common.MaxArgsCompletionFilter(1),
		),
	}
	config.AddCommand(useContext)

	useProject := &cobra.Command{
		Use:               "use-project [project-id]",
		Short:             "Set the project to query for project-scoped resources",
		Args:              cobra.MaximumNArgs(1),
		PersistentPreRunE: s.MakeInvokePreRunE(initCmdWScope),
		RunE:              cli.CtxWrap(cmdWScope.useProject),
		ValidArgsFunction: common.Complete(cli.HackCtxWrapCompletion(cmdWScope.useProjectCompletion, s),
			common.MaxArgsCompletionFilter(1)),
	}
	config.AddCommand(useProject)

	useEnvironment := &cobra.Command{
		Use:               "use-environment [environment-id]",
		Short:             "Set the environment to query for environment-scoped resources",
		Args:              cobra.MaximumNArgs(1),
		PersistentPreRunE: s.MakeInvokePreRunE(initCmdWScope),
		RunE:              cli.CtxWrap(cmdWScope.useEnvironment),
		ValidArgsFunction: common.Complete(cli.HackCtxWrapCompletion(cmdWScope.useEnvironmentCompletion, s),
			common.MaxArgsCompletionFilter(1)),
	}
	config.AddCommand(useEnvironment)

	currentContext := &cobra.Command{
		Use:               "current-context",
		Short:             "Display the current context",
		Args:              cobra.NoArgs,
		PersistentPreRunE: s.MakeInvokePreRunE(initCmdWScope),
		RunE:              cmdWScope.currentContext,
		Annotations: map[string]string{
			auth.OmitUser: "",
		},
	}
	config.AddCommand(currentContext)

	viewConfig := &cobra.Command{
		Use:               "view",
		Short:             "View Config",
		Args:              cobra.NoArgs,
		RunE:              cmdNoScope.viewConfig,
		PersistentPreRunE: s.MakeInvokePreRunE(initCmdNoScope),
		Annotations: map[string]string{
			auth.OmitUser: "",
		},
	}
	viewConfig.Flags().BoolVarP(&minify, "minify", "m", false,
		"Remove all information not used by current-context from the output")
	config.AddCommand(viewConfig)

	listConfig := &cobra.Command{
		Use:               "list-contexts",
		Short:             "list contexts",
		Args:              cobra.NoArgs,
		RunE:              cmdNoScope.listContexts,
		PersistentPreRunE: s.MakeInvokePreRunE(initCmdNoScope),
		Annotations: map[string]string{
			auth.OmitUser: "",
		},
	}
	config.AddCommand(listConfig)

	editConfig := &cobra.Command{
		Use:               "edit [context]",
		Short:             "Edit a context",
		Args:              cobra.MaximumNArgs(1),
		PersistentPreRunE: s.MakeInvokePreRunE(initCmdNoScope),
		RunE:              cmdNoScope.editConfig,
		ValidArgsFunction: common.Complete(
			cli.HackWrapCompletion(cmdNoScope.completions, s),
			common.MaxArgsCompletionFilter(1),
		),
		Annotations: map[string]string{
			auth.OmitUser: "",
		},
	}
	editConfig.Flags().StringVarP(&field, "field", "f", "", "Field to edit")

	if err := editConfig.RegisterFlagCompletionFunc("field",
		func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			options := []string{"name", "server"}
			labels := []string{}
			for _, option := range options {
				if strings.HasPrefix(option, toComplete) {
					labels = append(labels, option)
				}
			}
			return labels, cobra.ShellCompDirectiveNoFileComp
		}); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	editConfig.Flags().StringVarP(&value, "value", "v", "", "Value to set")
	editConfig.MarkFlagsRequiredTogether("field", "value")
	config.AddCommand(editConfig)

	parent.AddCommand(config)
}

func (c *CmdNoScope) completions(
	cmd *cobra.Command,
	args []string,
	toComplete string,
	s *cli.SetupContext,
) ([]string, cobra.ShellCompDirective) {
	names := []string{}

	if err := s.ExecuteInvokes(cmd, args, initCmdWScope); err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	for _, ctx := range c.Cfg.Contexts {
		if strings.HasPrefix(ctx.Name, toComplete) {
			var isCurrent string
			if ctx.Name == c.Cfg.CurrentContextName {
				isCurrent = "*"
			}
			names = append(names, ctx.Name+isCurrent)
		}
	}

	if len(names) == 0 {
		return nil, cobra.ShellCompDirectiveError
	}

	return names, cobra.ShellCompDirectiveNoFileComp
}

func (c *CmdWScope) useProjectCompletion(
	ctx context.Context,
	cmd *cobra.Command,
	args []string,
	toComplete string,
	s *cli.SetupContext,
) ([]string, cobra.ShellCompDirective) {
	if err := s.ExecuteInvokes(cmd, args, initCmdWScope); err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var projectIDs []string

	if c.Scope.GetCurrentContext() == nil || c.Scope.GetCurrentContext().GetAuth() == nil {
		return nil, cobra.ShellCompDirectiveError
	}

	resp, err := c.Rig.Project().List(ctx, &connect.Request[project.ListRequest]{
		Msg: &project.ListRequest{},
	})
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	for _, p := range resp.Msg.GetProjects() {
		if strings.HasPrefix(p.GetProjectId(), toComplete) {
			projectIDs = append(projectIDs, formatProject(p))
		}
	}

	if len(projectIDs) == 0 {
		return nil, cobra.ShellCompDirectiveError
	}

	return projectIDs, cobra.ShellCompDirectiveNoFileComp
}

func formatProject(p *project.Project) string {
	age := "-"
	if p.GetCreatedAt().IsValid() {
		age = p.GetCreatedAt().AsTime().Format("2006-01-02 15:04:05")
	}

	return fmt.Sprintf("%v\t (ID: %v, Created At: %v)", p.GetName(), p.GetProjectId(), age)
}

func (c *CmdWScope) useEnvironmentCompletion(
	ctx context.Context,
	cmd *cobra.Command,
	args []string,
	toComplete string,
	s *cli.SetupContext,
) ([]string, cobra.ShellCompDirective) {
	if err := s.ExecuteInvokes(cmd, args, initCmdWScope); err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var environmentIDs []string

	if c.Scope.GetCurrentContext() == nil || c.Scope.GetCurrentContext().GetAuth() == nil {
		return nil, cobra.ShellCompDirectiveError
	}

	resp, err := c.Rig.Environment().List(ctx, &connect.Request[environment.ListRequest]{
		Msg: &environment.ListRequest{},
	})
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	for _, p := range resp.Msg.GetEnvironments() {
		if strings.HasPrefix(p.GetEnvironmentId(), toComplete) {
			environmentIDs = append(environmentIDs, formatEnvironment(p))
		}
	}

	if len(environmentIDs) == 0 {
		return nil, cobra.ShellCompDirectiveError
	}

	return environmentIDs, cobra.ShellCompDirectiveNoFileComp
}

func formatEnvironment(e *environment.Environment) string {
	operatorVersion := "-"
	if e.GetOperatorVersion() != "" {
		operatorVersion = e.GetOperatorVersion()
	}

	return fmt.Sprintf("%v\t (Operator Version: %v)", e.GetEnvironmentId(), operatorVersion)
}
