package instance

import (
	"fmt"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/cmd/common"
	cmd_capsule "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/spf13/cobra"
)

func (c Cmd) get(cmd *cobra.Command, args []string) error {
	ctx := c.Ctx
	resp, err := c.Rig.Capsule().ListInstances(ctx, &connect.Request[capsule.ListInstancesRequest]{
		Msg: &capsule.ListInstancesRequest{
			CapsuleId: cmd_capsule.CapsuleID,
			Pagination: &model.Pagination{
				Offset: uint32(offset),
				Limit:  uint32(limit),
			},
		},
	})
	if err != nil {
		return err
	}

	instances := resp.Msg.GetInstances()

	if len(args) > 0 {
		found := false
		for _, i := range instances {
			if i.GetInstanceId() == args[0] {
				instances = []*capsule.Instance{i}
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("instance %q not found", args[0])
		}
	}

	if outputJSON {
		for _, i := range instances {
			cmd.Println(common.ProtoToPrettyJson(i))
		}
		return nil
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{fmt.Sprintf("Instances (%d)", resp.Msg.GetTotal()), "Build", "State", "Created At", "Uptime", "Restart Count"})
	for _, i := range instances {
		uptime := time.Since(i.GetStartedAt().AsTime())
		if i.GetFinishedAt().AsTime().After(i.GetStartedAt().AsTime()) {
			uptime = -time.Since(i.GetFinishedAt().AsTime())
		}
		t.AppendRow(table.Row{
			i.GetInstanceId(),
			i.GetBuildId(),
			i.GetState(),
			i.GetCreatedAt().AsTime().Format(time.RFC3339),
			uptime,
			i.GetRestartCount(),
		})
	}
	cmd.Println(t.Render())

	return nil
}
