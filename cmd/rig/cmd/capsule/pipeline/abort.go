package pipeline

import (
	"context"
	"strconv"

	"connectrpc.com/connect"
	capsule_api "github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
)

func (c *Cmd) abort(ctx context.Context, cmd *cobra.Command, args []string) error {
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

	_, err = c.Rig.Capsule().AbortPipeline(ctx, connect.NewRequest(&capsule_api.AbortPipelineRequest{
		ExecutionId: pipelineID,
	}))
	if err != nil {
		return err
	}

	cmd.Printf("pipeline %v aborted\n", pipelineID)
	return nil
}
