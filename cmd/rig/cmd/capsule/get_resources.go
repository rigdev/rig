package capsule

import (
	"context"
	"fmt"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/spf13/cobra"
)

func setupGetResources(parent *cobra.Command) {
	getResources := &cobra.Command{
		Use:   "get-resources",
		Short: "Displays the resource (container size) of the capsule",
		Args:  cobra.MaximumNArgs(1),
		RunE:  base.Register(GetResources),
	}
	getResources.Flags().BoolVar(&outputJSON, "json", false, "output as json")

	parent.AddCommand(getResources)
}

func GetResources(ctx context.Context, cmd *cobra.Command, capsuleID CapsuleID, client rig.Client) error {
	containerSettings, err := getCurrentContainerSettings(ctx, capsuleID, client)
	if err != nil {
		return err
	}
	if containerSettings == nil {
		fmt.Println("Capsule has no rollouts yet")
		return nil
	}

	if outputJSON {
		cmd.Println(common.ProtoToPrettyJson(containerSettings.Resources))
		return nil
	}

	limits := containerSettings.Resources.Limits
	requests := containerSettings.Resources.Requests

	t := table.NewWriter()
	t.AppendRows([]table.Row{{"", "Requests", "Limits"}})
	t.AppendSeparator()
	t.AppendRows([]table.Row{
		{"CPU", milliIntToString(uint64(requests.Cpu)), milliIntToString(uint64(limits.Cpu))},
		{"Memory", intToByteString(requests.Memory), intToByteString(limits.Memory)},
		{"Ephemeral Storage", intToByteString(requests.EphemeralStorage), intToByteString(limits.EphemeralStorage)},
	})
	cmd.Println(t.Render())

	return nil
}
