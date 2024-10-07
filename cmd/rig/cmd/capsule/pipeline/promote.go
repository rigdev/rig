package pipeline

import (
	"context"
	"fmt"
	"strconv"

	"connectrpc.com/connect"
	capsule_api "github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/cmd/common"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
)

type pipelineDryOutput struct {
	environment string
	out         capsule_cmd.DryOutput
}

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

	resp, err := c.Rig.Capsule().PromotePipeline(ctx, connect.NewRequest(&capsule_api.PromotePipelineRequest{
		ExecutionId: pipelineID,
		DryRun:      dryRun,
	}))
	if err != nil {
		return err
	}

	if !dryRun {
		cmd.Printf("pipeline execution %v progressed to phase %v \n", pipelineID,
			resp.Msg.GetStatus().GetPhaseStatuses()[len(resp.Msg.GetStatus().GetPhaseStatuses())-1].GetEnvironmentId())
		return nil
	}

	var envLabels []string
	var outs []*pipelineDryOutput
	for _, out := range resp.Msg.GetDryRunOutcomes() {

		fmt.Printf("Dry run for environment %v: %v\n", out.GetEnvironmentId(), out.GetOutcome())

		if out.GetOutcome() == nil {
			continue
		}

		out2, err := capsule_cmd.ProcessDryRunOutput(out.GetOutcome(), out.GetOutcome())
		if err != nil {
			return err
		}

		envLabels = append(envLabels, out.GetEnvironmentId())
		outs = append(outs, &pipelineDryOutput{
			environment: out.GetEnvironmentId(),
			out:         out2,
		})
	}

	if !c.Scope.IsInteractive() {
		outputType := flags.Flags.OutputType
		if outputType == common.OutputTypePretty {
			outputType = common.OutputTypeYAML
		}
		return common.FormatPrint(outs, outputType)
	}

	for {
		i, _, err := c.Prompter.Select("Select the environment to view the dry run output (CTRL + C to cancel)",
			envLabels)
		if err != nil {
			if common.ErrIsAborted(err) {
				return nil
			}
			return err
		}

		if err := capsule_cmd.PromptDryOutput(ctx, outs[i].out, c.Scheme); common.ErrIsAborted(err) {
			continue
		} else if err != nil {
			return err
		}
	}
}
