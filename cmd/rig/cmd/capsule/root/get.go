package root

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
)

func (c *Cmd) get(ctx context.Context, cmd *cobra.Command, _ []string) error {
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

	type output struct {
		Capsule *capsule.Capsule `json:"capsule" yaml:"capsule"`
		Rollout *capsule.Rollout `json:"rollout" yaml:"rollout"`
	}
	if base.Flags.OutputType != base.OutputTypePretty {
		var res []output
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

			res = append(res, output{
				Capsule: cc,
				Rollout: r.Msg.GetRollout(),
			})
		}

		if capsule_cmd.CapsuleID != "" {
			return base.FormatPrint(res[0])
		}
		return base.FormatPrint(res)
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
