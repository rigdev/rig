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

func get(ctx context.Context, cmd *cobra.Command, rc rig.Client, args []string) error {
	resp, err := rc.Capsule().List(ctx, &connect.Request[capsule.ListRequest]{
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

	capsules := resp.Msg.GetCapsules()

	if CapsuleID != "" {
		found := false
		for _, c := range resp.Msg.GetCapsules() {
			if c.GetCapsuleId() == CapsuleID {
				capsules = []*capsule.Capsule{c}
				break
			}
		}
		if !found {
			return errors.NotFoundErrorf("capsule %s not found", CapsuleID)
		}
	}

	if outputJSON {
		for _, c := range capsules {
			r, err := rc.Capsule().GetRollout(ctx, &connect.Request[capsule.GetRolloutRequest]{
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

			cmd.Println(common.ProtoToPrettyJson(c))
			if r.Msg.GetRollout() != nil {
				cmd.Println(common.ProtoToPrettyJson(r.Msg.GetRollout()))
			}
		}
		return nil
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{fmt.Sprintf("Capsules (%d)", resp.Msg.GetTotal()), "Replicas", "Build ID"})
	for _, c := range capsules {
		r, err := rc.Capsule().GetRollout(ctx, &connect.Request[capsule.GetRolloutRequest]{
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
			c.GetCapsuleId(),
			r.Msg.GetRollout().GetConfig().GetReplicas(),
			r.Msg.GetRollout().GetConfig().GetBuildId(),
		})
	}
	cmd.Println(t.Render())

	return nil
}
