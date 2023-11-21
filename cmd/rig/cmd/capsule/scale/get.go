package scale

import (
	"context"

	"github.com/jedib0t/go-pretty/v6/table"
	capsule_api "github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/spf13/cobra"
)

func (c *Cmd) get(ctx context.Context, cmd *cobra.Command, _ []string) error {
	rollout, err := capsule.GetCurrentRollout(ctx, c.Rig)
	if err != nil {
		return err
	}
	containerSettings, replicas, err := capsule.GetCurrentContainerResources(ctx, c.Rig)
	if err != nil {
		return err
	}

	if base.Flags.OutputType != base.OutputTypePretty {
		obj := scaleObj{
			Replicas:      rollout.GetConfig().GetReplicas(),
			ContainerSize: containerSettings.GetResources(),
			Autoscaler:    rollout.GetConfig().GetHorizontalScale(),
		}
		return base.FormatPrint(obj)
	}

	limits := containerSettings.GetResources().GetLimits()
	requests := containerSettings.GetResources().GetRequests()

	t := table.NewWriter()
	t.AppendRow(table.Row{"", "Requests", "Limits"})
	t.AppendSeparator()
	t.AppendRows([]table.Row{
		{
			"CPU",
			milliIntToString(uint64(requests.GetCpuMillis())),
			formatLimitString(milliIntToString, uint64(limits.GetCpuMillis())),
		},
		{
			"Memory",
			intToByteString(requests.GetMemoryBytes()),
			formatLimitString(intToByteString, limits.GetMemoryBytes()),
		},
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

type scaleObj struct {
	Replicas      uint32                       `json:"replicas"`
	ContainerSize *capsule_api.Resources       `json:"resources"`
	Autoscaler    *capsule_api.HorizontalScale `json:"autoscaler,omitempty"`
}
