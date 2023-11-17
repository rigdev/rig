package service_account

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-api/api/v1/service_account"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
)

func (c Cmd) list(ctx context.Context, cmd *cobra.Command, args []string) error {
	resp, err := c.Rig.ServiceAccount().List(ctx, &connect.Request[service_account.ListRequest]{
		Msg: &service_account.ListRequest{},
	})
	if err != nil {
		return err
	}

	serviceAccounts := resp.Msg.GetServiceAccounts()

	if len(args) > 0 {
		found := false
		for _, c := range resp.Msg.GetServiceAccounts() {
			if c.GetServiceAccountId() == args[0] {
				serviceAccounts = []*service_account.Entry{c}
				found = true
				break
			}
		}
		if !found {
			return errors.NotFoundErrorf("service account %s not found", args[0])
		}
	}

	if outputJSON {
		for _, cred := range serviceAccounts {
			cmd.Println(common.ProtoToPrettyJson(cred))
		}
		return nil
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{"Service Accounts", "Name", "ID", "ClientID"})
	for i, cred := range serviceAccounts {
		t.AppendRow(table.Row{i + 1, cred.GetServiceAccount().GetName(), cred.GetServiceAccountId(), cred.GetClientId()})
	}

	cmd.Println(t.Render())

	return nil
}
