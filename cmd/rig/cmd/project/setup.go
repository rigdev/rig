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
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/rigdev/rig/cmd/rig/cmd/cmdconfig"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var (
	offset int
	limit  int
)

var (
	name  string
	field string
	value string
)

var useProject bool

type Cmd struct {
	fx.In

	Rig rig.Client
	Cfg *cmdconfig.Config
}

var cmd Cmd

func initCmd(c Cmd) {
	cmd.Rig = c.Rig
	cmd.Cfg = c.Cfg
	fmt.Println("project", cmd.Cfg.GetProject())
}

func Setup(parent *cobra.Command) {
	project := &cobra.Command{
		Use:               "project",
		Short:             "Manage Rig projects",
		PersistentPreRunE: base.MakeInvokePreRunE(initCmd),
	}

	getSettings := &cobra.Command{
		Use:   "get-settings",
		Short: "Get settings for the current project",
		Args:  cobra.NoArgs,
		RunE:  base.CtxWrap(cmd.getSettings),
	}
	project.AddCommand(getSettings)

	updateSettings := &cobra.Command{
		Use:   "update-settings",
		Short: "Update settings for the current project",
		Args:  cobra.NoArgs,
		RunE:  base.CtxWrap(cmd.updateSettings),
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
		Use:   "create",
		Short: "Create a new project",
		Args:  cobra.NoArgs,
		RunE:  base.CtxWrap(cmd.create),
		Annotations: map[string]string{
			base.OmitProject: "",
		},
	}
	createProject.Flags().StringVarP(&name, "name", "n", "", "Project name")
	createProject.Flags().BoolVar(&useProject, "use", false, "Use the created project")
	if err := createProject.RegisterFlagCompletionFunc("use", common.BoolCompletions); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	project.AddCommand(createProject)

	deleteProject := &cobra.Command{
		Use:   "delete",
		Short: "Delete the current project",
		Args:  cobra.NoArgs,
		RunE:  base.CtxWrap(cmd.delete),
	}
	project.AddCommand(deleteProject)

	getProject := &cobra.Command{
		Use:   "get ",
		Short: "Get the current project",
		Args:  cobra.NoArgs,
		RunE:  base.CtxWrap(cmd.get),
	}
	project.AddCommand(getProject)

	updateProject := &cobra.Command{
		Use:   "update",
		Short: "Update the current project",
		Args:  cobra.NoArgs,
		RunE:  base.CtxWrap(cmd.update),
	}
	updateProject.Flags().StringVarP(&field, "field", "f", "", "Field to update")
	updateProject.Flags().StringVarP(&value, "value", "v", "", "Value to set")
	updateProject.MarkFlagsRequiredTogether("field", "value")
	updateProject.SetHelpFunc(
		func(cmd *cobra.Command, args []string) {
			cmd.Printf(
				("Usage:\n" +
					"  update [flags] \n\n" +
					"Flags:\n" +
					"  -f, --field string   Field to update\n" +
					"  -h, --help           help for update\n" +
					"  -v, --value string   Value to set\n" +

					"Avaliable fields:\n" +
					"  name - string \n"),
			)
		},
	)
	if err := updateProject.RegisterFlagCompletionFunc("field", projectUpdateFieldsCompletion); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	project.AddCommand(updateProject)

	listProjects := &cobra.Command{
		Use:   "list",
		Short: "List projects",
		Args:  cobra.NoArgs,
		RunE:  base.CtxWrap(cmd.list),
		Annotations: map[string]string{
			base.OmitProject: "",
		},
	}
	listProjects.Flags().IntVar(&offset, "offset", 0, "Offset")
	listProjects.Flags().IntVarP(&limit, "limit", "l", 10, "Limit")
	project.AddCommand(listProjects)

	use := &cobra.Command{
		Use:   "use [project-id | project-name]",
		Short: "Set the project to query for project-scoped resources",
		Args:  cobra.MaximumNArgs(1),
		RunE:  base.CtxWrap(cmd.use),
		Annotations: map[string]string{
			base.OmitProject: "",
		},
		ValidArgsFunction: base.CtxWrapCompletion(cmd.useProjectCompletion),
	}
	project.AddCommand(use)

	parent.AddCommand(project)
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

func projectUpdateFieldsCompletion(
	_ *cobra.Command,
	_ []string,
	toComplete string,
) ([]string, cobra.ShellCompDirective) {
	fields := []string{"name"}
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

func (c *Cmd) useProjectCompletion(
	ctx context.Context,
	_ *cobra.Command,
	_ []string,
	toComplete string,
) ([]string, cobra.ShellCompDirective) {
	var projectIDs []string

	if c.Cfg.GetCurrentContext() == nil || c.Cfg.GetCurrentAuth() == nil {
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
