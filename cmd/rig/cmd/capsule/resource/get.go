package resource

import (
	"context"
	"fmt"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/spf13/cobra"
)

func get(ctx context.Context, cmd *cobra.Command, client rig.Client) error {
	containerSettings, replicas, err := capsule.GetCurrentContainerResources(ctx, client)
	if err != nil {
		return err
	}
	if containerSettings == nil {
		fmt.Println("Capsule has no rollouts yet")
		return nil
	}

	if outputJSON {
		cmd.Println(common.ProtoToPrettyJson(containerSettings.Resources))
		cmd.Println("{replicas: ", replicas, "}")
		return nil
	}

	limits := containerSettings.Resources.Limits
	requests := containerSettings.Resources.Requests

	t := table.NewWriter()
	t.AppendRow(table.Row{"", "Requests", "Limits"})
	t.AppendSeparator()
	t.AppendRows([]table.Row{
		{"CPU", milliIntToString(uint64(requests.CpuMillis)), formatLimitString(milliIntToString, uint64(limits.CpuMillis))},
		{"Memory", intToByteString(requests.MemoryBytes), formatLimitString(intToByteString, limits.MemoryBytes)},
	})
	t.AppendSeparator()
	t.AppendRow(table.Row{"Replicas", replicas})
	cmd.Println(t.Render())

	return nil
}

func formatLimitString(fmt func(uint64) string, n uint64) string {
	if n == 0 {
		return "-"
	}
	return fmt(n)
}
