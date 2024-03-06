package plugins

import (
	"errors"

	"github.com/rigdev/rig/cmd/rig-ops/cmd/base"
	"github.com/spf13/cobra"
)

var (
	capsules   []string
	namespaces []string
	plugins    []string

	showConfig bool

	specPath     string
	pluginConfig string
	output       string
	replaces     []string
	removes      []int
	appends      []string
	dry          bool
	interactive  bool
)

func Setup(parent *cobra.Command) {
	pluginsCmd := &cobra.Command{
		Use:   "plugins",
		Short: "Migrate you kubernetes deployments to Rig Capsules",
	}

	check := &cobra.Command{
		Use:   "check",
		Short: "Check which plugins will be run on which capsules",
		RunE:  base.Register(check),
	}
	//nolint:lll
	check.Flags().StringSliceVar(&capsules, "capsules", nil, "If given, will use those capsule names instead of reading them from the platform")
	check.Flags().StringSliceVar(&namespaces, "namespaces", nil, "If given, will only use those namespaces")
	check.Flags().StringSliceVar(&plugins, "plugins", nil, "If given, will only use those plugins names.")
	pluginsCmd.AddCommand(check)

	listSteps := &cobra.Command{
		Use:   "list-steps",
		Short: "Lists the plugin steps currently configured in the operator",
		RunE:  base.Register(listSteps),
	}
	//nolint:lll
	listSteps.Flags().BoolVar(&showConfig, "show-config", false, "If set, will also display the YAML configuration for each plugin.")
	pluginsCmd.AddCommand(listSteps)

	get := &cobra.Command{
		Use: "get 2",
		//nolint:lll
		Short: "Gets the configuration for a single plugin given by index. If no index is given, it will prompt you to choose a plugin.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  base.Register(get),
	}
	pluginsCmd.AddCommand(get)

	dryRun := &cobra.Command{
		Use: "dry-run namespace my-capsule",
		//nolint:lll
		Short: "runs a dry-run of the operator on the given namespace and capsule (or provided capsule spec)",
		Long: `Runs a dry-run of the operator on the given namespace and capsule (or provided capsule spec).
Besides giving a complete list of plugin configurations, is possible to edit the plugin configuration
for the dry-run using the replace, remove and append flags.
If any of these are given, first all the replace, then all remove then all append commands will be executed.
The dry run will be executed with the resulting list of plugins.`,
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) != 0 && len(args) != 2 {
				return errors.New("takes exactly 0 or 2 arguments")
			}
			return nil
		},
		RunE: base.Register(dryRun),
	}
	//nolint:lll
	dryRun.Flags().StringVar(&pluginConfig, "config", "", "If given, will read the config file at the path and use that as the plugin config. If empty, will use the plugin config of the running operator.")
	//nolint:lll
	dryRun.Flags().StringVar(&specPath, "spec", "", "If given, will read the capsule spec at the path instead of using the capsule spec of an existing capsule from the platform")
	//nolint:lll
	dryRun.Flags().StringSliceVar(&replaces, "replace", nil, "Assumes argument of the form '<idx>:<path>' (e.g. '2:conf.yaml'). Will replace the plugin at the given index (0-indexed) with the config at the path.")
	//nolint:lll
	dryRun.Flags().IntSliceVar(&removes, "remove", nil, "Will remove the plugins at the specified index(es) (0-indexed)")
	//nolint:lll
	dryRun.Flags().StringSliceVar(&appends, "append", nil, "Will append plugins given by the configs at the given paths. Will append them in the order of the arguments.")
	//nolint:lll
	dryRun.Flags().BoolVar(&dry, "dry", false, "If given, will only show the list of plugins used for the dry-run")
	//nolint:lll
	dryRun.Flags().StringVar(&output, "output-path", "", "If given, will write the output to a file at the given path.")
	dryRun.Flags().BoolVar(&interactive, "interactive", false, "If set, will show differences interactively.")
	pluginsCmd.AddCommand(dryRun)

	list := &cobra.Command{
		Use:   "list",
		Short: "Lists the set of plugins available in the operator",
		RunE:  base.Register(list),
	}
	pluginsCmd.AddCommand(list)

	parent.AddCommand(pluginsCmd)
}
