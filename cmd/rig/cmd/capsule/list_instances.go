package capsule

import (
	"context"
	"fmt"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
)

func CapsuleListInstances(ctx context.Context, cmd *cobra.Command, capsuleID CapsuleID, nc rig.Client) error {
	resp, err := nc.Capsule().ListInstances(ctx, &connect.Request[capsule.ListInstancesRequest]{
		Msg: &capsule.ListInstancesRequest{
			CapsuleId: capsuleID,
			Pagination: &model.Pagination{
				Offset: uint32(offset),
				Limit:  uint32(limit),
			},
		},
	})
	if err != nil {
		return err
	}

	if outputJSON {
		for _, i := range resp.Msg.GetInstances() {
			cmd.Println(common.ProtoToPrettyJson(i))
		}
		return nil
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{fmt.Sprintf("Instances (%d)", resp.Msg.GetTotal()), "Build", "State", "Created At", "Uptime", "Restart Count"})
	for _, i := range resp.Msg.GetInstances() {
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
