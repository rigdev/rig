package config

import (
	"context"
	"fmt"

	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/rigdev/rig/cmd/rig/cmd/utils"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func ConfigInit(ctx context.Context, cmd *cobra.Command, cfg *base.Config, logger *zap.Logger) error {
	if ok, err := utils.PromptConfirm("Do you want to configure a new context", true); err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("aborted")
	}

	return base.CreateContext(cfg)
}
