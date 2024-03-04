package plugins

import (
	"github.com/rigdev/rig/cmd/rig-ops/cmd/base"
	"github.com/spf13/cobra"
)

var (
	operatorConfig string
	capsules       []string
	projects       []string
	environments   []string
	plugins        []string

	showConfig bool
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
	check.Flags().StringVar(&operatorConfig, "operator-config", "", "If given, will read the config file at the path and use that as an operator config. If empty, will use the operator config of the running operator.")
	//nolint:lll
	check.Flags().StringSliceVar(&capsules, "capsules", nil, "If given, will use those capsule names instead of reading them from the platform")
	//nolint:lll
	check.Flags().StringSliceVar(&projects, "projects", nil, "If given, will use those project names instead of reading them from the platform")
	//nolint:lll
	check.Flags().StringSliceVar(&environments, "environments", nil, "If given, will use those environment names instead of reading them from the platform. The environments given must be known to the platform.")
	//nolint:lll
	check.Flags().StringSliceVar(&plugins, "plugins", nil, "If given, will only use those plugins names.")
	pluginsCmd.AddCommand(check)

	list := &cobra.Command{
		Use:   "list",
		Short: "Lists the plugins currently configured in the operator",
		RunE:  base.Register(list),
	}
	list.Flags().BoolVar(&showConfig, "show-config", false, "If set, will also display the YAML configuration for each plugin.")
	pluginsCmd.AddCommand(list)

	get := &cobra.Command{
		Use:   "get 2",
		Short: "Gets the configuration for a single plugin given by index. If no index is given, it will prompt you to choose a plugin.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  base.Register(get),
	}
	pluginsCmd.AddCommand(get)

	parent.AddCommand(pluginsCmd)
}
