package rollout

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
)

func (c *Cmd) get(ctx context.Context, cmd *cobra.Command, args []string) error {
	resp, err := c.Rig.Capsule().ListRollouts(ctx, &connect.Request[capsule.ListRolloutsRequest]{
		Msg: &capsule.ListRolloutsRequest{
			CapsuleId: capsule_cmd.CapsuleID,
			Pagination: &model.Pagination{
				Offset:     uint32(offset),
				Limit:      uint32(limit),
				Descending: true,
			},
			ProjectId:     c.Cfg.GetProject(),
			EnvironmentId: base.Flags.Environment,
		},
	})
	if err != nil {
		return err
	}

	rollouts := resp.Msg.GetRollouts()
	if len(args) > 0 {
		found := false
		for _, r := range resp.Msg.GetRollouts() {
			id, err := strconv.ParseUint(args[0], 10, 32)
			if err != nil {
				return errors.InvalidArgumentErrorf("invalid rollout id - %v", err)
			}
			if r.GetRolloutId() == id {
				rollouts = []*capsule.Rollout{r}
				found = true
				break
			}
		}
		if !found {
			return errors.NotFoundErrorf("rollout %s not found", args[0])
		}
	}

	if base.Flags.OutputType != base.OutputTypePretty {
		return base.FormatPrint(rollouts)
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{
		fmt.Sprintf("Rollouts (%d)", resp.Msg.GetTotal()),
		"Deployed At",
		"Replicas",
		"State",
		"Created By",
	})
	for i, r := range rollouts {
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
