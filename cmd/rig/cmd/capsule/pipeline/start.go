package pipeline

import (
	"context"

	"connectrpc.com/connect"
	capsule_api "github.com/rigdev/rig-go-api/api/v1/capsule"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
)

func (c *Cmd) start(ctx context.Context, cmd *cobra.Command, args []string) error {
	pipelineName := ""
	var err error
	if len(args) < 2 {
		if !c.Scope.IsInteractive() {
			return errors.InvalidArgumentErrorf("missing pipeline name")
		}

		pipelineName, err = c.promptForPipelineName(ctx)
		if err != nil {
			return err
		}
	} else {
		pipelineName = args[1]
	}

	resp, err := c.Rig.Capsule().StartPipeline(ctx, connect.NewRequest(&capsule_api.StartPipelineRequest{
		ProjectId:    c.Scope.GetCurrentContext().GetProject(),
		CapsuleId:    capsule_cmd.CapsuleID,
		PipelineName: pipelineName,
	}))
	if err != nil {
		return err
	}

	cmd.Printf("pipeline %v started with execution id %v\n", pipelineName, resp.Msg.Status.GetExecutionId())

	return nil
}
