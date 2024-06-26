package serviceaccount

import (
	"context"

	"connectrpc.com/connect"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-api/api/v1/service_account"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/spf13/cobra"
)

func (c *Cmd) list(ctx context.Context, cmd *cobra.Command, _ []string) error {
	resp, err := c.Rig.ServiceAccount().List(ctx, &connect.Request[service_account.ListRequest]{
		Msg: &service_account.ListRequest{},
	})
	if err != nil {
		return err
	}

	serviceAccounts := resp.Msg.GetServiceAccounts()

	if flags.Flags.OutputType != common.OutputTypePretty {
		return common.FormatPrint(serviceAccounts, flags.Flags.OutputType)
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{"Service Accounts", "Name", "ID", "ClientID"})
	for i, cred := range serviceAccounts {
		t.AppendRow(table.Row{i + 1, cred.GetName(), cred.GetServiceAccountId(), cred.GetClientId()})
	}

	cmd.Println(t.Render())

	return nil
}
