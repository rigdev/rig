package pipeline

import (
	"context"
	"strconv"

	"connectrpc.com/connect"
	capsule_api "github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/cmd/common"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
)

func (c *Cmd) progress(ctx context.Context, cmd *cobra.Command, args []string) error {
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

	resp, err := c.Rig.Capsule().ProgressPipeline(ctx, connect.NewRequest(&capsule_api.ProgressPipelineRequest{
		ExecutionId: pipelineID,
		DryRun:      dryRun,
	}))
	if err != nil {
		return err
	}

	if !dryRun {
		cmd.Printf("pipeline execution %v progressed to phase %v", pipelineID,
			resp.Msg.GetStatus().GetPhaseStatuses()[len(resp.Msg.GetStatus().GetPhaseStatuses())-1].GetEnvironmentId())
		return nil
	}

	out := capsule_cmd.ProcessDryRunOutput(resp.Msg.GetOutcome(), resp.Msg.GetRevision().GetSpec(), c.Scheme)

	if !c.Scope.IsInteractive() {
		outputType := flags.Flags.OutputType
		if outputType == common.OutputTypePretty {
			outputType = common.OutputTypeYAML
		}
		return common.FormatPrint(out, outputType)
	}

	return capsule_cmd.PromptDryOutput(ctx, out, resp.Msg.GetOutcome())
}
