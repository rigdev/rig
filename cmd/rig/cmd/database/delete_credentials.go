package database

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/rig/cmd/utils"
	"github.com/spf13/cobra"
)

func DeleteCredentials(ctx context.Context, cmd *cobra.Command, args []string, nc rig.Client) error {
	identifier := ""
	if len(args) > 0 {
		identifier = args[0]
	}
	_, id, err := utils.GetDatabase(ctx, identifier, nc)
	if err != nil {
		return err
	}

	if clientID == "" {
		clientID, err = utils.PromptGetInput("Client ID", utils.ValidateNonEmpty)
		if err != nil {
			return err
		}
	}

	if _, err := nc.Database().DeleteCredentials(ctx, &connect.Request[database.DeleteCredentialsRequest]{
		Msg: &database.DeleteCredentialsRequest{
			DatabaseId: id,
			ClientId:   clientID,
		},
	}); err != nil {
		return err
	}

	cmd.Printf("deleted credential %s\n", name)
	return nil
}
