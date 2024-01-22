package project

import (
	"context"

	"connectrpc.com/connect"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-api/api/v1/project/settings"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/spf13/cobra"
)

func (c *Cmd) getSettings(ctx context.Context, cmd *cobra.Command, _ []string) error {
	req := &settings.GetSettingsRequest{}
	resp, err := c.Rig.ProjectSettings().GetSettings(ctx, &connect.Request[settings.GetSettingsRequest]{Msg: req})
	if err != nil {
		return err
	}
	set := resp.Msg.GetSettings()

	if flags.Flags.OutputType != common.OutputTypePretty {
		return common.FormatPrint(set, flags.Flags.OutputType)
	}

	dockerRegistries := []table.Row{}
	for i, r := range set.GetDockerRegistries() {
		if i == 0 {
			dockerRegistries = append(dockerRegistries, table.Row{"Docker Registries", r})
			continue
		}
		dockerRegistries = append(dockerRegistries, table.Row{"", r.GetHost()})
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{"Attribute", "Value"})
	t.AppendRows(dockerRegistries)

	cmd.Println(t.Render())
	return nil
}
