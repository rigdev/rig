package root

import (
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/cmd/common"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
)

func (c Cmd) get(cmd *cobra.Command, args []string) error {
	ctx := c.Ctx
	resp, err := c.Rig.Capsule().List(ctx, &connect.Request[capsule.ListRequest]{
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

	if capsule_cmd.CapsuleID != "" {
		found := false
		for _, c := range resp.Msg.GetCapsules() {
			if c.GetCapsuleId() == capsule_cmd.CapsuleID {
				capsules = []*capsule.Capsule{c}
				found = true
				break
			}
		}
		if !found {
			return errors.NotFoundErrorf("capsule %s not found", capsule_cmd.CapsuleID)
		}
	}

	if outputJSON {
		for _, cc := range capsules {
			r, err := c.Rig.Capsule().GetRollout(ctx, &connect.Request[capsule.GetRolloutRequest]{
				Msg: &capsule.GetRolloutRequest{
					CapsuleId: cc.GetCapsuleId(),
					RolloutId: cc.GetCurrentRollout(),
				},
			})
			if errors.IsNotFound(err) {
				// OK, default values.
				r = &connect.Response[capsule.GetRolloutResponse]{}
			} else if err != nil {
				return err
			}

			cmd.Println(common.ProtoToPrettyJson(cc))
			if r.Msg.GetRollout() != nil {
				cmd.Println(common.ProtoToPrettyJson(r.Msg.GetRollout()))
			}
		}
		return nil
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{fmt.Sprintf("Capsules (%d)", resp.Msg.GetTotal()), "Replicas", "Build ID"})
	for _, cc := range capsules {
		r, err := c.Rig.Capsule().GetRollout(ctx, &connect.Request[capsule.GetRolloutRequest]{
			Msg: &capsule.GetRolloutRequest{
				CapsuleId: cc.GetCapsuleId(),
				RolloutId: cc.GetCurrentRollout(),
			},
		})
		if errors.IsNotFound(err) {
			// OK, default values.
			r = &connect.Response[capsule.GetRolloutResponse]{}
		} else if err != nil {
			return err
		}

		t.AppendRow(table.Row{
			cc.GetCapsuleId(),
			r.Msg.GetRollout().GetConfig().GetReplicas(),
			r.Msg.GetRollout().GetConfig().GetBuildId(),
		})
	}
	cmd.Println(t.Render())

	return nil
}
