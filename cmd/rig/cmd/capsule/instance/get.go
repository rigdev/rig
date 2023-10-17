package instance

import (
	"encoding/json"
	"fmt"

	// "strings"

	"github.com/bufbuild/connect-go"
	// "github.com/fatih/color"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	// "github.com/rigdev/rig-go-api/api/v1/capsule/instance"
	"github.com/rigdev/rig-go-api/model"
	cmd_capsule "github.com/rigdev/rig/cmd/rig/cmd/capsule"

	// table2 "github.com/rodaine/table"
	"github.com/spf13/cobra"
)

func (c Cmd) get(cmd *cobra.Command, args []string) error {
	ctx := c.Ctx

	// TODO Fix to use the new dataformat
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
	//
	// headerFmt := color.New(color.FgBlue, color.Underline).SprintfFunc()
	// columnFmt := color.New(color.FgYellow).SprintfFunc()
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

// func instanceStatusToTableRows(instance *instance.Status) [][]any {
// 	rowLength := len(instance.GetStateMachine().GetStates()) + 1
// 	rows := [][]any{
// 		make([]any, rowLength),
// 		make([]any, rowLength),
// 	}
//
// 	rows[0][0] = instance.GetInstanceId()
// 	for idx, s := range instance.GetStateMachine().GetStates() {
// 		m := s.GetSubStateMachines()[0]
// 		s = m.GetStates()[0]
// 		str := s.GetStateId().String()
// 		str, _ = strings.CutPrefix(str, "STATE_ID_")
// 		rows[0][idx+1] = str
// 		rows[1][idx+1] = s.GetMessage()
// 	}
//
// 	return rows
// }
