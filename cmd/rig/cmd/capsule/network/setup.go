package network

import (
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/spf13/cobra"
)

var (
	outputJSON bool
)

func Setup(parent *cobra.Command) *cobra.Command {
	network := &cobra.Command{
		Use:   "network",
		Short: "Configure and inspect the network of the capsule",
	}

	networkConfigure := &cobra.Command{
		Use:   "configure [network-file]",
		Short: "configure the network of the capsule",
		Args:  cobra.MaximumNArgs(1),
		RunE:  base.Register(configure),
	}

	network.AddCommand(networkConfigure)

	networkGet := &cobra.Command{
		Use:               "get [name]",
		Short:             "get the entire network or a specific interface of the capsule",
		Args:              cobra.MaximumNArgs(1),
		RunE:              base.Register(get),
		ValidArgsFunction: common.Complete(capsule.NetworkCompletions, common.MaxArgsCompletionFilter(1)),
	}
	networkGet.Flags().BoolVar(&outputJSON, "json", false, "output as json")
	networkGet.RegisterFlagCompletionFunc("json", common.BoolCompletions)
	network.AddCommand(networkGet)

	parent.AddCommand(network)

	return network
}
