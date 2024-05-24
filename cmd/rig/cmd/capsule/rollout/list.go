package rollout

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/cmd/common"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	table2 "github.com/rodaine/table"
	"github.com/spf13/cobra"
)

func (c *Cmd) list(ctx context.Context, _ *cobra.Command, _ []string) error {
	headerFmt := color.New(color.FgBlue, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()
	tbl := table2.New("ID", "Deployed At", "Replicas", "State", "Created By")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	if !follow {
		resp, err := c.Rig.Capsule().ListRollouts(ctx, &connect.Request[capsule.ListRolloutsRequest]{
			Msg: &capsule.ListRolloutsRequest{
				CapsuleId: capsule_cmd.CapsuleID,
				Pagination: &model.Pagination{
					Offset:     uint32(offset),
					Limit:      uint32(limit),
					Descending: true,
				},
				ProjectId:     flags.GetProject(c.Scope),
				EnvironmentId: flags.GetEnvironment(c.Scope),
			},
		})
		if err != nil {
			return err
		}
		rollouts := resp.Msg.GetRollouts()

		if flags.Flags.OutputType != common.OutputTypePretty {
			return common.FormatPrint(rollouts, flags.Flags.OutputType)
		}

		var rows [][]string
		for _, r := range rollouts {
			rows = append(rows, rolloutToRow(r))
		}
		tbl.SetRows(rows)
		tbl.Print()
		return nil
	}

	stream, err := c.Rig.Capsule().WatchRollouts(ctx, connect.NewRequest(&capsule.WatchRolloutsRequest{
		CapsuleId:     capsule_cmd.CapsuleID,
		ProjectId:     flags.GetProject(c.Scope),
		EnvironmentId: flags.GetEnvironment(c.Scope),
		Pagination: &model.Pagination{
			Offset:     uint32(offset),
			Limit:      uint32(limit),
			Descending: true,
		},
	}))
	if err != nil {
		return err
	}

	defer stream.Close()

	var lock sync.Mutex
	rollouts := make(map[uint64]*capsule.Rollout)
	shouldPrint := false
	highestRolloutID := uint64(0)
	go func() {
		time.Sleep(1 * time.Second)
		for {
			if !shouldPrint {
				continue
			}

			lock.Lock()
			rows := make([][]string, len(rollouts))
			idDiff := highestRolloutID - uint64(len(rollouts))
			for i, r := range rollouts {
				rows[i-idDiff-1] = rolloutToRow(r)
			}

			if len(rows) == 0 {
				lock.Unlock()
				continue
			}

			tbl.SetRows(rows)
			tbl.Print()
			shouldPrint = false
			lock.Unlock()
			time.Sleep(3 * time.Second)
		}
	}()

	for stream.Receive() {
		lock.Lock()
		rollout := stream.Msg().GetUpdated()
		rollouts[rollout.GetRolloutId()] = rollout
		if rollout.GetRolloutId() > highestRolloutID {
			highestRolloutID = rollout.GetRolloutId()
		}
		shouldPrint = true
		lock.Unlock()
	}

	if stream.Err() != nil {
		return err
	}

	return nil
}

func rolloutToRow(r *capsule.Rollout) []string {
	return []string{
		fmt.Sprint(r.GetRolloutId()),
		common.FormatTime(r.GetConfig().GetCreatedAt().AsTime()),
		fmt.Sprint(r.GetConfig().GetReplicas()),
		strings.TrimPrefix(r.GetStatus().GetState().String(), "STATE_"),
		r.GetConfig().GetCreatedBy().GetPrintableName(),
	}
}
