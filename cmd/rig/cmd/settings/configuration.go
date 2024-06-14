package settings

import (
	"context"

	"connectrpc.com/connect"
	settings_api "github.com/rigdev/rig-go-api/api/v1/settings"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/spf13/cobra"
)

func (c *Cmd) configuration(ctx context.Context, _ *cobra.Command, _ []string) error {
	resp, err := c.Rig.Settings().GetConfiguration(ctx, connect.NewRequest(&settings_api.GetConfigurationRequest{}))
	if err != nil {
		return err
	}

	if flags.Flags.OutputType != common.OutputTypePretty {
		return common.FormatPrint(resp.Msg.GetConfiguration(), flags.Flags.OutputType)
	}

	return common.FormatPrint(resp.Msg.GetConfiguration(), common.OutputTypeYAML)
}
