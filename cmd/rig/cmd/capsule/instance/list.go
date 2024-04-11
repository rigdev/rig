package instance

import (
	"context"
	"strings"

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
	resp, err := c.Rig.Capsule().ListInstanceStatuses(ctx, connect.NewRequest(&capsule.ListInstanceStatusesRequest{
		CapsuleId: cmd_capsule.CapsuleID,
		Pagination: &model.Pagination{
			Offset:     uint32(offset),
			Limit:      uint32(limit),
			Descending: true,
		},
		ProjectId:       flags.GetProject(c.Scope),
		EnvironmentId:   flags.GetEnvironment(c.Scope),
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

	headerFmt := color.New(color.FgBlue, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()
	tbl := table2.New("ID", "Created", "Deleted", "Scheduling", "Preparing", "Running", "Deleted")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)
	for _, i := range instances {
		tbl.AddRow(instanceStatusToTableRow(i)...)
	}
	tbl.Print()

	return nil
}

func instanceStatusToTableRow(instance *instance.Status) []any {
	stages := instance.GetStages()
	d := stages.GetDeleted().GetInfo().GetUpdatedAt()
	ds := "-"
	if d != nil {
		ds = common.FormatTime(d.AsTime())
	}
	return []any{
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
