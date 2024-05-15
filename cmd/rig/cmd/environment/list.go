package environment

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-api/api/v1/environment"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/spf13/cobra"
)

func (c *Cmd) list(ctx context.Context, cmd *cobra.Command, _ []string) error {
	req := &environment.ListRequest{
		ExcludeEphemeral: ephemeral,
		ProjectFilter:    projectFilter,
	}
	resp, err := c.Rig.Environment().List(ctx, &connect.Request[environment.ListRequest]{Msg: req})
	if err != nil {
		return err
	}

	if flags.Flags.OutputType != common.OutputTypePretty {
		return common.FormatPrint(resp.Msg.GetEnvironments(), flags.Flags.OutputType)
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{
		fmt.Sprintf("Environments (%d)", len(resp.Msg.GetEnvironments())),
		"Name",
		"Cluster",
		"Namespace Template",
		"Ephemeral",
		"Global",
		"Active Projects",
	})
	for i, p := range resp.Msg.GetEnvironments() {
		activeProjects := "None"
		if len(p.GetActiveProjects()) > 0 {
			activeProjects = strings.Join(p.GetActiveProjects(), ", ")
		}
		t.AppendRow(table.Row{
			i + 1,
			p.GetEnvironmentId(),
			p.GetClusterId(),
			p.GetNamespaceTemplate(),
			p.GetEphemeral(),
			p.GetGlobal(),
			activeProjects,
		})
	}
	cmd.Println(t.Render())
	return nil
}
