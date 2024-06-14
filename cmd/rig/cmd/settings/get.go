package settings

import (
	"context"

	"connectrpc.com/connect"
	settings_api "github.com/rigdev/rig-go-api/api/v1/settings"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/spf13/cobra"
)

func (c *Cmd) get(ctx context.Context, cmd *cobra.Command, _ []string) error {
	resp, err := c.Rig.Settings().GetSettings(ctx, connect.NewRequest(&settings_api.GetSettingsRequest{}))
	if err != nil {
		return err
	}

	if resp.Msg.GetSettings() == nil {
		cmd.Println("No settings set")
		return nil
	}

	if flags.Flags.OutputType != common.OutputTypePretty {
		return common.FormatPrint(resp.Msg.GetSettings(), flags.Flags.OutputType)
	}

	return common.FormatPrint(resp.Msg.GetSettings(), common.OutputTypeYAML)
}
