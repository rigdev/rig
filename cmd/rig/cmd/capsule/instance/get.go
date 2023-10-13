package instance

import (
	"encoding/json"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/model"
	cmd_capsule "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/spf13/cobra"
)

func (c Cmd) get(cmd *cobra.Command, args []string) error {
	ctx := c.Ctx
	resp, err := c.Rig.Capsule().ListAllCurrentInstanceStatuses(ctx, connect.NewRequest(&capsule.ListAllCurrentInstanceStatusesRequest{
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

	// TODO use the new InstanceStatus data format and fix this code

	// headerFmt := color.New(color.FgBlue, color.Underline).SprintfFunc()
	// columnFmt := color.New(color.FgYellow).SprintfFunc()
	//
	// tbl := table2.New("ID", "Scheduling", "Preparing", "Running")
	// tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)
	// for _, i := range instances {
	// 	for _, r := range instanceStatusToTableRows(i) {
	// 		tbl.AddRow(r...)
	// 	}
	// }
	// tbl.Print()

	return nil
}

// func instanceStatusToTableRows(instance *capsule.InstanceStatus) [][]any {
// 	rows := [][]any{
// 		{"", "", "", ""},
// 		{"", "", "", ""},
// 	}
// 	rows[0][0] = instance.GetData().GetInstanceId()
//
// 	schedule := instance.GetStages().GetSchedule()
// 	rows[1][1] = schedule.GetMessage()
//
// 	pulling := instance.GetStages().GetPreparing().GetStages().GetPulling()
// 	rows[1][2] = pulling.GetMessage()
//
// 	running := instance.GetStages().GetRunning()
// 	if crashLoop := running.GetStages().GetCrashLoopBackoff(); crashLoop != nil {
// 		rows[0][3] = "CRASH_LOOP"
// 		rows[1][3] = crashLoop.GetMessage()
// 	}
// 	if ready := running.GetStages().GetReady(); ready != nil {
// 		rows[1][3] = running.GetMessage()
// 	}
// 	if running := running.GetStages().GetRunning(); running != nil {
// 		rows[0][3] = "RUNNING"
// 	}
//
// 	return rows
// }
