package service_account

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-api/api/v1/service_account"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/rig/cmd/utils"
	"github.com/spf13/cobra"
)

func ServiceAccountList(ctx context.Context, cmd *cobra.Command, args []string, nc rig.Client) error {
	resp, err := nc.ServiceAccount().List(ctx, &connect.Request[service_account.ListRequest]{
		Msg: &service_account.ListRequest{},
	})
	if err != nil {
		return err
	}

	if outputJSON {
		for _, cred := range resp.Msg.GetServiceAccounts() {
			cmd.Println(utils.ProtoToPrettyJson(cred))
		}
		return nil
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{"Service Accounts", "Name", "ID", "ClientID"})
	for i, cred := range resp.Msg.GetServiceAccounts() {
		t.AppendRow(table.Row{i + 1, cred.GetServiceAccount().GetName(), cred.GetServiceAccountId(), cred.GetClientId()})
	}

	cmd.Println(t.Render())

	return nil
}
