package instance

import (
	"context"
	"strings"

	"connectrpc.com/connect"
	"github.com/fatih/color"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/api/v1/capsule/instance"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	cmd_capsule "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	table2 "github.com/rodaine/table"
	"github.com/spf13/cobra"
)

func (c *Cmd) get(ctx context.Context, _ *cobra.Command, _ []string) error {
	resp, err := c.Rig.Capsule().ListInstanceStatuses(ctx, connect.NewRequest(&capsule.ListInstanceStatusesRequest{
		CapsuleId: cmd_capsule.CapsuleID,
		Pagination: &model.Pagination{
			Offset: uint32(offset),
			Limit:  uint32(limit),
		},
		ProjectId:       c.Cfg.GetProject(),
		EnvironmentId:   base.GetEnvironment(c.Cfg),
		ExcludeExisting: excludeExisting,
		IncludeDeleted:  includeDeleted,
	}))
	if err != nil {
		return err
	}
	instances := resp.Msg.GetInstances()

	if base.Flags.OutputType != base.OutputTypePretty {
		return base.FormatPrint(instances)
	}

	headerFmt := color.New(color.FgBlue, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()
	tbl := table2.New("ID", "Scheduling", "Preparing", "Running", "Deleted")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)
	for _, i := range instances {
		tbl.AddRow(instanceStatusToTableRow(i)...)
	}
	tbl.Print()

	return nil
}

func instanceStatusToTableRow(instance *instance.Status) []any {
	stages := instance.GetStages()
	return []any{
		instance.GetInstanceId(),
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
