package instance

import (
	"context"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/fatih/color"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/api/v1/capsule/instance"
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

	headerFmt := color.New(color.FgBlue, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()
	tbl := table2.New("ID", "Created", "Deleted", "Scheduling", "Preparing", "Running", "Deleted")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	if !follow {
		resp, err := c.Rig.Capsule().GetInstanceStatus(ctx, connect.NewRequest[capsule.GetInstanceStatusRequest](
			&capsule.GetInstanceStatusRequest{
				CapsuleId:     capsule_cmd.CapsuleID,
				InstanceId:    instanceID,
				ProjectId:     c.Scope.GetCurrentContext().GetProject(),
				EnvironmentId: c.Scope.GetCurrentContext().GetEnvironment(),
			},
		))
		if err != nil {
			return err
		}

		if flags.Flags.OutputType != common.OutputTypePretty {
			return common.FormatPrint(resp.Msg.GetStatus(), flags.Flags.OutputType)
		}
		tbl.SetRows([][]string{
			instanceStatusToTableRow(resp.Msg.GetStatus()),
		})
		tbl.Print()
		return nil
	}

	stream, err := c.Rig.Capsule().WatchInstanceStatuses(ctx, connect.NewRequest(&capsule.WatchInstanceStatusesRequest{
		CapsuleId:      capsule_cmd.CapsuleID,
		ProjectId:      c.Scope.GetCurrentContext().GetProject(),
		EnvironmentId:  c.Scope.GetCurrentContext().GetEnvironment(),
		InstanceId:     instanceID,
		IncludeDeleted: true,
	}))
	if err != nil {
		return err
	}

	defer stream.Close()

	shouldPrint := false
	var lock sync.Mutex
	var status *instance.Status

	go func() {
		for {
			if !shouldPrint || status == nil {
				continue
			}
			lock.Lock()
			tbl.SetRows([][]string{
				instanceStatusToTableRow(status),
			})
			tbl.Print()
			shouldPrint = false
			lock.Unlock()
			time.Sleep(2 * time.Second)
		}
	}()

	for stream.Receive() {
		if stream.Msg().GetUpdated() == nil {
			continue
		}

		lock.Lock()
		status = stream.Msg().GetUpdated()

		if flags.Flags.OutputType != common.OutputTypePretty {
			common.FormatPrint(status, flags.Flags.OutputType)
			continue
		}

		shouldPrint = true
		lock.Unlock()

	}

	return nil
}
