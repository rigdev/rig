package instance

import (
	"encoding/json"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/fatih/color"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/model"
	cmd_capsule "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	table2 "github.com/rodaine/table"
	"github.com/spf13/cobra"
)

func (c Cmd) get(cmd *cobra.Command, args []string) error {
	ctx := c.Ctx
	resp, err := c.Rig.Capsule().ListInstanceStatuses(ctx, connect.NewRequest(&capsule.ListInstanceStatusesRequest{
		CapsuleId: cmd_capsule.CapsuleID,
		Pagination: &model.Pagination{
			Offset: uint32(offset),
			Limit:  uint32(limit),
		},
	}))
	if err != nil {
		return err
	}
	instances := resp.Msg.GetInstances()

	if outputJSON {
		jsonStr, err := json.MarshalIndent(instances, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(jsonStr))
		return nil
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

func instanceStatusToTableRows(instance *capsule.InstanceStatus) [][]any {
	rows := [][]any{
		{"", "", "", ""},
		{"", "", "", ""},
	}
	rows[0][0] = "id"

	schedule := instance.GetSchedule()
	rows[0][1] = capsule.ScheduleState_name[int32(schedule.GetState())]
	rows[1][1] = schedule.GetMessage()

	pulling := instance.GetPreparing().GetPulling()
	rows[0][2] = capsule.ImagePullingState_name[int32(pulling.GetState())]
	rows[1][2] = pulling.GetMessage()

	running := instance.GetRunning()
	if crashLoop := running.GetCrashLoopBackoff(); crashLoop != nil {
		rows[0][3] = "CRASH_LOOP"
		rows[1][3] = crashLoop.GetMessage()
	}
	if ready := running.GetReady(); ready != nil {
		rows[0][3] = capsule.InstanceRunningReadyState_name[int32(ready.GetState())]
		rows[1][3] = running.GetMessage()
	}
	if running := running.GetRunning(); running != nil {
		rows[0][3] = "RUNNING"
	}

	return rows
}
