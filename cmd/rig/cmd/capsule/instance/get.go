package instance

import (
	"context"

	"connectrpc.com/connect"
	"github.com/fatih/color"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/cmd/common"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	table2 "github.com/rodaine/table"
	"github.com/spf13/cobra"
)

func (c *Cmd) get(ctx context.Context, cmd *cobra.Command, args []string) error {
	arg := ""
	if len(args) > 1 {
		arg = args[1]
	}

	instanceID, err := c.provideInstanceID(ctx, capsule_cmd.CapsuleID, arg, cmd.ArgsLenAtDash())
	if err != nil {
		return err
	}

	resp, err := c.Rig.Capsule().GetInstanceStatus(ctx, connect.NewRequest[capsule.GetInstanceStatusRequest](
		&capsule.GetInstanceStatusRequest{
			CapsuleId:     capsule_cmd.CapsuleID,
			InstanceId:    instanceID,
			ProjectId:     flags.GetProject(c.Scope),
			EnvironmentId: flags.GetEnvironment(c.Scope),
		},
	))
	if err != nil {
		return err
	}

	if flags.Flags.OutputType != common.OutputTypePretty {
		return common.FormatPrint(resp.Msg.GetStatus(), flags.Flags.OutputType)
	}

	headerFmt := color.New(color.FgBlue, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()
	tbl := table2.New("ID", "Created", "Deleted", "Scheduling", "Preparing", "Running", "Deleted")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)
	tbl.AddRow(instanceStatusToTableRow(resp.Msg.GetStatus())...)
	tbl.Print()

	return nil
}
