package project

import (
	"context"

	"github.com/spf13/cobra"
)

func (c *Cmd) create(ctx context.Context, _ *cobra.Command, args []string) error {
	var id string
	if len(args) > 0 {
		id = args[0]
	}
	return c.Auth.CreateProject(ctx, id, &useProject)
}
