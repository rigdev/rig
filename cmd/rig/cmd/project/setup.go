package project

import (
	"context"
	"fmt"
	"os"
	"strings"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/project"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/services/auth"
	"github.com/rigdev/rig/pkg/cli"
	"github.com/rigdev/rig/pkg/cli/scope"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var (
	offset int
	limit  int
)

var (
	field string
	value string
)

var (
	useProject bool
	current    bool
)

type Cmd struct {
	fx.In

	Rig      rig.Client
	Scope    scope.Scope
	Auth     *auth.Service
	Prompter common.Prompter
}

var cmd Cmd

func initCmd(c Cmd) {
	cmd = c
}

func Setup(parent *cobra.Command, s *cli.SetupContext) {
	project := &cobra.Command{
		Use:               "project",
		Short:             "Manage Rig projects",
		PersistentPreRunE: s.MakeInvokePreRunE(initCmd),
		Annotations: map[string]string{
			auth.OmitEnvironment: "",
			auth.OmitProject:     "",
		},
		GroupID: common.ManagementGroupID,
	}

	getSettings := &cobra.Command{
		Use:   "get-settings",
		Short: "Get project settings",
		Args:  cobra.NoArgs,
		RunE:  cli.CtxWrap(cmd.getSettings),
	}
	project.AddCommand(getSettings)

	updateSettings := &cobra.Command{
		Use:   "update-settings",
		Short: "Update project settings",
		Args:  cobra.NoArgs,
		RunE:  cli.CtxWrap(cmd.updateSettings),
	}
	updateSettings.Flags().StringVarP(&field, "field", "f", "", "Field to update")
	updateSettings.Flags().StringVarP(&value, "value", "v", "", "Value to set")
	updateSettings.MarkFlagsRequiredTogether("field", "value")
	updateSettings.SetHelpFunc(
		func(cmd *cobra.Command, args []string) {
			cmd.Printf(
				("Usage:\n" +
					"  update-settings [flags] \n\n" +
					"Flags:\n" +
					"  -f, --field string   Field to update\n" +
					"  -h, --help           help for update-settings\n" +
					"  -v, --value string   Value to set\n" +

					"Avaliable fields:\n" +
					"  email-provder - json \n" +
					"  add-docker-registry - json \n" +
					"  delete-docker-registry - string \n" +
					"  template - json \n"),
			)
		},
	)
	if err := updateSettings.RegisterFlagCompletionFunc("field", settingsUpdateFieldsCompletion); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	project.AddCommand(updateSettings)

	createProject := &cobra.Command{
		Use:   "create [project-id]",
		Short: "Create a new project",
		Args:  cobra.MaximumNArgs(1),
		RunE:  cli.CtxWrap(cmd.create),
	}
	createProject.Flags().BoolVar(&useProject, "use", false, "Use the created project")
	if err := createProject.RegisterFlagCompletionFunc("use", common.BoolCompletions); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	project.AddCommand(createProject)

	deleteProject := &cobra.Command{
		Use:   "delete [project-id]",
		Short: "Delete a project. If project-ID is left out, delete the current project",
		Args:  cobra.MaximumNArgs(1),
		ValidArgsFunction: common.Complete(
			cli.HackCtxWrapCompletion(cmd.projectCompletions, s),
			common.MaxArgsCompletionFilter(1),
		),
		RunE: cli.CtxWrap(cmd.delete),
	}
	project.AddCommand(deleteProject)

	getProjects := &cobra.Command{
		Use:   "list",
		Short: "Get one or multiple projects",
		Args:  cobra.MaximumNArgs(1),
		RunE:  cli.CtxWrap(cmd.list),
	}
	getProjects.Flags().IntVar(&offset, "offset", 0, "Offset")
	getProjects.Flags().IntVarP(&limit, "limit", "l", 10, "Limit")
	project.AddCommand(getProjects)

	parent.AddCommand(project)
}

func (c *Cmd) projectCompletions(ctx context.Context,
	cmd *cobra.Command,
	args []string,
	toComplete string,
	s *cli.SetupContext,
) ([]string, cobra.ShellCompDirective) {
	if current {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	if err := s.ExecuteInvokes(cmd, args, initCmd); err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	completions := []string{}
	projects, err := c.Rig.Project().List(ctx, &connect.Request[project.ListRequest]{
		Msg: &project.ListRequest{},
	})
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	for _, p := range projects.Msg.Projects {
		if strings.HasPrefix(p.GetProjectId(), toComplete) {
			completions = append(completions, formatProject(p))
		}
	}

	if len(completions) == 0 {
		return nil, cobra.ShellCompDirectiveError
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

func settingsUpdateFieldsCompletion(
	_ *cobra.Command,
	_ []string,
	toComplete string,
) ([]string, cobra.ShellCompDirective) {
	fields := []string{"email-provider", "add-docker-registry", "delete-docker-registry", "template"}
	var completions []string
	for _, s := range fields {
		if strings.HasPrefix(s, toComplete) {
			completions = append(completions, s)
		}
	}
	if len(completions) == 0 {
		return nil, cobra.ShellCompDirectiveError
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

func formatProject(p *project.Project) string {
	return fmt.Sprintf("%s\t (Created At: %v)", p.GetProjectId(), p.GetCreatedAt().AsTime().Format("2006-01-02 15:04:05"))
}
