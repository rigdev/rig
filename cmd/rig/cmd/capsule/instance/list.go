package instance

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/fatih/color"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/api/v1/capsule/instance"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/cmd/common"
	cmd_capsule "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	table2 "github.com/rodaine/table"
	"github.com/spf13/cobra"
)

func (c *Cmd) list(ctx context.Context, _ *cobra.Command, _ []string) error {
	headerFmt := color.New(color.FgBlue, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()
	tbl := table2.New("ID", "Created", "Deleted", "Scheduling", "Preparing", "Running", "Deleted")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	if !follow {
		resp, err := c.Rig.Capsule().ListInstanceStatuses(ctx, connect.NewRequest(&capsule.ListInstanceStatusesRequest{
			CapsuleId: cmd_capsule.CapsuleID,
			Pagination: &model.Pagination{
				Offset:     uint32(offset),
				Limit:      uint32(limit),
				Descending: true,
			},
			ProjectId:       c.Scope.GetCurrentContext().GetProject(),
			EnvironmentId:   c.Scope.GetCurrentContext().GetEnvironment(),
			ExcludeExisting: excludeExisting,
			IncludeDeleted:  includeDeleted,
		}))
		if err != nil {
			return err
		}
		instances := resp.Msg.GetInstances()

		if flags.Flags.OutputType != common.OutputTypePretty {
			return common.FormatPrint(instances, flags.Flags.OutputType)
		}

		var rows [][]string
		for _, i := range instances {
			rows = append(rows, instanceStatusToTableRow(i))
		}
		tbl.SetRows(rows)
		tbl.Print()
		return nil
	}

	stream, err := c.Rig.Capsule().WatchInstanceStatuses(ctx, connect.NewRequest(&capsule.WatchInstanceStatusesRequest{
		CapsuleId:       cmd_capsule.CapsuleID,
		ProjectId:       c.Scope.GetCurrentContext().GetProject(),
		EnvironmentId:   c.Scope.GetCurrentContext().GetEnvironment(),
		ExcludeExisting: excludeExisting,
		IncludeDeleted:  includeDeleted,
	}))
	if err != nil {
		return err
	}

	defer stream.Close()

	var lock sync.Mutex
	shouldPrint := true
	statuses := make(map[string]*instance.Status)

	go func() {
		for {
			if !shouldPrint {
				continue
			}
			lock.Lock()
			var instances []*instance.Status
			for _, i := range statuses {
				instances = append(instances, i)
			}

			sort.Slice(instances, func(i, j int) bool {
				iIsDeleted := instances[i].GetStages().GetDeleted() != nil
				jIsDeleted := instances[j].GetStages().GetDeleted() != nil
				if iIsDeleted && !jIsDeleted {
					return true
				}
				if !iIsDeleted && jIsDeleted {
					return false
				}
				return instances[i].CreatedAt.AsTime().Before(instances[j].CreatedAt.AsTime())
			})

			var rows [][]string
			for _, i := range instances {
				rows = append(rows, instanceStatusToTableRow(i))
			}

			tbl.SetRows(rows)
			tbl.Print()
			shouldPrint = false
			lock.Unlock()
			time.Sleep(3 * time.Second)
		}
	}()

	for stream.Receive() {
		instanceStatus := stream.Msg().GetResponse()
		switch v := instanceStatus.(type) {
		case *capsule.WatchInstanceStatusesResponse_Updated:
			lock.Lock()
			statuses[v.Updated.InstanceId] = v.Updated
			shouldPrint = true
			lock.Unlock()
		case *capsule.WatchInstanceStatusesResponse_Deleted:
			if includeDeleted {
				continue
			}
			lock.Lock()
			delete(statuses, v.Deleted)
			shouldPrint = true
			lock.Unlock()
		}
	}

	if stream.Err() != nil {
		return err
	}

	return nil
}

func instanceStatusToTableRow(instance *instance.Status) []string {
	stages := instance.GetStages()
	d := stages.GetDeleted().GetInfo().GetUpdatedAt()
	ds := "-"
	if d != nil {
		ds = common.FormatTime(d.AsTime())
	}
	return []string{
		instance.GetInstanceId(),
		common.FormatTime(instance.CreatedAt.AsTime()),
		ds,
		formatRow(stages.GetSchedule()),
		formatRow(stages.GetPreparing()),
		formatRow(stages.GetRunning()),
		formatRow(stages.GetDeleted()),
	}
}

type stage interface {
	GetInfo() *instance.StageInfo
}

func formatRow(stage stage) string {
	info := stage.GetInfo()
	if info.GetState() == instance.StageState_STAGE_STATE_UNSPECIFIED {
		return ""
	}
	return formatStageState(info.GetState())
}

func formatStageState(s instance.StageState) string {
	if s == instance.StageState_STAGE_STATE_UNSPECIFIED {
		return ""
	}
	return strings.ToLower(strings.TrimPrefix(s.String(), "STAGE_STATE_"))
}
