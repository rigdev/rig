package serviceaccount

import (
	"context"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/service_account"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/pkg/uuid"
	"github.com/spf13/cobra"
)

func (c *Cmd) delete(ctx context.Context, cmd *cobra.Command, args []string) error {
	var serviceAccountID string
	var err error

	if len(args) > 0 {
		serviceAccountID = args[0]
	}

	if serviceAccountID == "" {
		serviceAccountID, err = c.Prompter.Input("Service Account ID:", common.ValidateNonEmptyOpt)
		if err != nil {
			return err
		}
	}

	_, err = uuid.Parse(serviceAccountID)
	if err != nil {
		return err
	}

	_, err = c.Rig.ServiceAccount().Delete(ctx, &connect.Request[service_account.DeleteRequest]{
		Msg: &service_account.DeleteRequest{
			ServiceAccountId: serviceAccountID,
		},
	})
	if err != nil {
		return err
	}

	cmd.Println("Credential deleted")
	return nil
}
