package capsule

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
)

func CapsuleList(ctx context.Context, cmd *cobra.Command, nc rig.Client, args []string) error {
	resp, err := nc.Capsule().List(ctx, &connect.Request[capsule.ListRequest]{
		Msg: &capsule.ListRequest{
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
		for _, c := range resp.Msg.GetCapsules() {
			cmd.Println(common.ProtoToPrettyJson(c))
		}
		return nil
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{fmt.Sprintf("Capsules (%d)", resp.Msg.GetTotal()), "ID", "Replicas", "Build ID"})
	for _, c := range resp.Msg.GetCapsules() {
		r, err := nc.Capsule().GetRollout(ctx, &connect.Request[capsule.GetRolloutRequest]{
			Msg: &capsule.GetRolloutRequest{
				CapsuleId: c.GetCapsuleId(),
				RolloutId: c.GetCurrentRollout(),
			},
		})
		if errors.IsNotFound(err) {
			// OK, default values.
			r = &connect.Response[capsule.GetRolloutResponse]{}
		} else if err != nil {
			return err
		}

		t.AppendRow(table.Row{
			c.GetName(),
			c.GetCapsuleId(),
			r.Msg.GetRollout().GetConfig().GetReplicas(),
			r.Msg.GetRollout().GetConfig().GetBuildId(),
		})
	}
	cmd.Println(t.Render())

	return nil
}
