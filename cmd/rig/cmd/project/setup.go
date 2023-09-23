package project

import (
	"context"

	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/rigdev/rig/cmd/rig/cmd/cmd_config"
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

var (
	outputJSON bool
	useProject bool
)

type Cmd struct {
	fx.In

	Ctx context.Context
	Rig rig.Client
	Cfg *cmd_config.Config
}

func (c Cmd) Setup(parent *cobra.Command) {
	project := &cobra.Command{
		Use:   "project",
		Short: "Manage Rig projects",
	}

	getSettings := &cobra.Command{
		Use:  "get-settings",
		Args: cobra.NoArgs,
		RunE: c.getSettings,
	}
	getSettings.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	project.AddCommand(getSettings)

	updateSettings := &cobra.Command{
		Use:  "update-settings",
		Args: cobra.NoArgs,
		RunE: c.updateSettings,
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
	project.AddCommand(updateSettings)

	createProject := &cobra.Command{
		Use:  "create",
		Args: cobra.NoArgs,
		RunE: c.create,
		Annotations: map[string]string{
			base.OmitProject: "",
		},
	}
	createProject.Flags().StringVarP(&name, "name", "n", "", "Project name")
	createProject.Flags().BoolVar(&useProject, "use", false, "Use the created project")
	project.AddCommand(createProject)

	deleteProject := &cobra.Command{
		Use:  "delete",
		Args: cobra.NoArgs,
		RunE: c.delete,
	}
	project.AddCommand(deleteProject)

	getProject := &cobra.Command{
		Use:  "get ",
		Args: cobra.NoArgs,
		RunE: c.get,
	}
	getProject.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	project.AddCommand(getProject)

	updateProject := &cobra.Command{
		Use:  "update",
		Args: cobra.NoArgs,
		RunE: c.update,
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
	project.AddCommand(updateProject)

	listProjects := &cobra.Command{
		Use:  "list",
		Args: cobra.NoArgs,
		RunE: c.list,
		Annotations: map[string]string{
			base.OmitProject: "",
		},
	}
	listProjects.Flags().IntVarP(&offset, "offset", "o", 0, "Offset")
	listProjects.Flags().IntVarP(&limit, "limit", "l", 10, "Limit")
	listProjects.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	project.AddCommand(listProjects)

	use := &cobra.Command{
		Use:   "use [project-id | project-name]",
		Short: "Set the project to query for project-scoped resources",
		Args:  cobra.MaximumNArgs(1),
		RunE:  c.use,
	}
	project.AddCommand(use)

	parent.AddCommand(project)
}
