package config

import (
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/spf13/cobra"
)

func Setup(parent *cobra.Command) {
	config := &cobra.Command{
		Use: "config",
	}

	init := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new context",
		Args:  cobra.NoArgs,
		RunE:  base.Register(ConfigInit),
		Annotations: map[string]string{
			base.OmitProject: "",
			base.OmitUser:    "",
		},
	}
	config.AddCommand(init)

	useContext := &cobra.Command{
		Use:   "use-context [context]",
		Short: "Change the current context to use",
		Args:  cobra.MaximumNArgs(1),
		RunE:  base.Register(ConfigUseContext),
		Annotations: map[string]string{
			base.OmitProject: "",
			base.OmitUser:    "",
		},
	}
	config.AddCommand(useContext)

	parent.AddCommand(config)
}
