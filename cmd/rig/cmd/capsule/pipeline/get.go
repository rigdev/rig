package pipeline

import (
	"context"
	"strconv"

	"connectrpc.com/connect"
	"github.com/fatih/color"
	capsule_api "github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/rigdev/rig/pkg/errors"
	table2 "github.com/rodaine/table"
	"github.com/spf13/cobra"
)

func (c *Cmd) get(ctx context.Context, _ *cobra.Command, args []string) error {
	pipelineIDStr := ""
	var err error
	if len(args) == 0 {
		if !c.Scope.IsInteractive() {
			return errors.InvalidArgumentErrorf("missing pipeline execution id")
		}

		pipelineIDStr, err = c.promptForPipelineID(ctx)
		if err != nil {
			return err
		}
	} else {
		pipelineIDStr = args[0]
	}

	pipelineID, err := strconv.ParseUint(pipelineIDStr, 10, 64)
	if err != nil {
		return err
	}

	resp, err := c.Rig.Capsule().GetPipelineStatus(ctx, connect.NewRequest(&capsule_api.GetPipelineStatusRequest{
		ExecutionId: pipelineID,
	}))
	if err != nil {
		return err
	}

	if flags.Flags.OutputType != common.OutputTypePretty {
		return common.FormatPrint(resp.Msg.GetStatus(), flags.Flags.OutputType)
	}

	headerFmt := color.New(color.FgBlue, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()
	tbl := table2.New("ID", "Pipeline", "Started", "State",
		"Current Phase", "Phase State", "Phase Rollout", "Phase Message")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	rows := [][]string{
		pipelineStatusToTableRow(resp.Msg.GetStatus()),
	}
	tbl.SetRows(rows)
	tbl.Print()

	return nil
}
