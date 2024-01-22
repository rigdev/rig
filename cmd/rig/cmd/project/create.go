package project

import (
	"context"

	"github.com/spf13/cobra"
)

func (c *Cmd) create(ctx context.Context, _ *cobra.Command, _ []string) error {
	return c.Auth.CreateProject(ctx, name, &useProject)
}
