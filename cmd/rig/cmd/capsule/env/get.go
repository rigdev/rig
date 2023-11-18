package env

import (
	"context"

	"github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/spf13/cobra"
)

func (c *Cmd) get(ctx context.Context, cmd *cobra.Command, args []string) error {
	r, err := capsule.GetCurrentRollout(ctx, c.Rig)
	if err != nil {
		return err
	}

	if r.GetConfig().GetContainerSettings().GetEnvironmentVariables() == nil {
		cmd.Println("No environment variables set")
	}

	if len(args) == 0 {
		for k, v := range r.GetConfig().GetContainerSettings().GetEnvironmentVariables() {
			cmd.Println(k, "=", v)
		}
		return nil
	}

	value := r.GetConfig().GetContainerSettings().GetEnvironmentVariables()[args[0]]
	if value == "" {
		cmd.Println("No environment variable set for key", args[0])
	}

	cmd.Println(value)
	return nil
}
