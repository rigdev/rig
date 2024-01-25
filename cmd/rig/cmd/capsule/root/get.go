package root

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/cmd/common"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
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
			ProjectId: flags.GetProject(c.Cfg),
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

	var outputs []output
	for _, cc := range capsules {
		r, err := capsule_cmd.GetCurrentRolloutOfCapsule(ctx, c.Rig, c.Cfg, cc.GetCapsuleId())
		if errors.IsNotFound(err) {
			// OK, default values.
			r = &capsule.Rollout{}
		} else if err != nil {
			return err
		}

		outputs = append(outputs, output{
			Capsule: cc,
			Rollout: r,
		})
	}

	if flags.Flags.OutputType != common.OutputTypePretty {
		return common.FormatPrint(outputs, flags.Flags.OutputType)
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{fmt.Sprintf("Capsules (%d)", resp.Msg.GetTotal()), "Replicas", "Build ID"})
	for _, o := range outputs {
		t.AppendRow(table.Row{
			o.Capsule.GetCapsuleId(),
			o.Rollout.GetConfig().GetReplicas(),
			o.Rollout.GetConfig().GetBuildId(),
		})
	}
	cmd.Println(t.Render())

	return nil
}
