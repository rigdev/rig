package capsule

import (
	"context"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig-go-sdk"
	"github.com/spf13/cobra"
)

func CapsuleEvents(ctx context.Context, cmd *cobra.Command, capsuleID CapsuleID, nc rig.Client) error {
	if rollout == 0 {
		resp, err := nc.Capsule().Get(ctx, &connect.Request[capsule.GetRequest]{
			Msg: &capsule.GetRequest{
				CapsuleId: capsuleID.String(),
			},
		})
		if err != nil {
			return err
		}

		rollout = resp.Msg.GetCapsule().GetCurrentRollout()
	}

	resp, err := nc.Capsule().ListEvents(ctx, &connect.Request[capsule.ListEventsRequest]{
		Msg: &capsule.ListEventsRequest{
			CapsuleId: capsuleID.String(),
			Pagination: &model.Pagination{
				Offset:     uint32(offset),
				Limit:      uint32(limit),
				Descending: true,
			},
			RolloutId: rollout,
		},
	})
	if err != nil {
		return err
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{"Created At", "Created By", "Message"})
	for _, e := range resp.Msg.GetEvents() {
		t.AppendRow(table.Row{
			e.GetCreatedAt().AsTime().Format(time.RFC822),
			e.GetCreatedBy().GetPrintableName(),
			e.GetMessage(),
		})
	}
	cmd.Println(t.Render())

	return nil
}
