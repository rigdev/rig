package capsule

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/rig/cmd/utils"
	"github.com/spf13/cobra"
)

func CapsuleListRollouts(ctx context.Context, cmd *cobra.Command, capsuleID CapsuleID, nc rig.Client) error {
	resp, err := nc.Capsule().ListRollouts(ctx, &connect.Request[capsule.ListRolloutsRequest]{
		Msg: &capsule.ListRolloutsRequest{
			CapsuleId: capsuleID.String(),
			Pagination: &model.Pagination{
				Offset:     uint32(offset),
				Limit:      uint32(limit),
				Descending: true,
			},
		},
	})
	if err != nil {
		return err
	}

	if outputJSON {
		for _, r := range resp.Msg.GetRollouts() {
			cmd.Println(utils.ProtoToPrettyJson(r))
		}
		return nil
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{fmt.Sprintf("Rollouts (%d)", resp.Msg.GetTotal()), "Deployed At", "Replicas", "State", "Created By"})
	for i, r := range resp.Msg.GetRollouts() {
		id := fmt.Sprint("#", r.GetRolloutId())
		if i == 0 {
			id = fmt.Sprint(id, " (current)")
		}

		t.AppendRow(table.Row{
			id,
			r.GetConfig().GetCreatedAt().AsTime().Format(time.RFC822),
			r.GetConfig().GetReplicas(),
			fmt.Sprint(
				strings.TrimPrefix(r.GetStatus().GetState().String(), "ROLLOUT_STATE_"),
				" - ",
				r.GetStatus().GetMessage(),
			),
			r.GetConfig().GetCreatedBy().GetPrintableName(),
		})
	}
	cmd.Println(t.Render())

	return nil
}
