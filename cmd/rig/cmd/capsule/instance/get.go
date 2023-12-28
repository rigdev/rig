package instance

import (
	"context"

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
	// TODO Fix to use the new dataformat
	resp, err := c.Rig.Capsule().ListInstanceStatuses(ctx, connect.NewRequest(&capsule.ListInstanceStatusesRequest{
		CapsuleId: cmd_capsule.CapsuleID,
		Pagination: &model.Pagination{
			Offset: uint32(offset),
			Limit:  uint32(limit),
		},
		ProjectId:     c.Cfg.GetProject(),
		EnvironmentId: base.Flags.Environment,
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
	tbl := table2.New("ID", "Scheduling", "Preparing", "Running")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)
	for _, i := range instances {
		for _, r := range instanceStatusToTableRows(i) {
			tbl.AddRow(r...)
		}
	}
	tbl.Print()

	return nil
}

func instanceStatusToTableRows(instance *instance.Status) [][]any {
	rowLength := 4
	rows := [][]any{
		make([]any, rowLength),
		make([]any, rowLength),
	}

	rows[0][0] = instance.GetInstanceId()
	rows[1][0] = ""
	stages := instance.GetStages()

	rows[0][1] = stages.GetSchedule().GetInfo().GetName()
	rows[1][1] = stages.GetSchedule().GetInfo().GetState().String()

	rows[0][2] = stages.GetPreparing().GetInfo().GetName()
	rows[1][2] = stages.GetPreparing().GetInfo().GetState().String()

	rows[0][3] = stages.GetRunning().GetInfo().GetName()
	rows[1][3] = stages.GetRunning().GetInfo().GetState().String()

	return rows
}
