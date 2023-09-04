package project

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-api/api/v1/project/settings"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/rig/cmd/utils"
	"github.com/spf13/cobra"
)

func ProjectGetSettings(ctx context.Context, cmd *cobra.Command, args []string, nc rig.Client) error {
	req := &settings.GetSettingsRequest{}
	resp, err := nc.ProjectSettings().GetSettings(ctx, &connect.Request[settings.GetSettingsRequest]{Msg: req})
	if err != nil {
		return err
	}
	set := resp.Msg.GetSettings()

	if outputJSON {
		cmd.Println(utils.ProtoToPrettyJson(set))
		return nil
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
	t.AppendRows([]table.Row{
		{"Email Provider", set.GetEmailProvider().GetInstance().String()},
		{" - Client ID", set.GetEmailProvider().GetClientId()},
		{" - From Email", set.GetEmailProvider().GetFrom()},
		{"Text Provider", set.GetTextProvider().GetInstance().String()},
		{" - Client ID", set.GetTextProvider().GetClientId()},
		{" - From Phone", set.GetTextProvider().GetFrom()},
	})
	t.AppendRows(dockerRegistries)

	cmd.Println(t.Render())
	return nil
}
