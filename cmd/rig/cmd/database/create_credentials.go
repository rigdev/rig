package database

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/rig/cmd/utils"
	"github.com/spf13/cobra"
)

func CreateCredentials(ctx context.Context, cmd *cobra.Command, args []string, nc rig.Client) error {
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
	res, err := nc.Database().CreateCredentials(ctx, &connect.Request[database.CreateCredentialsRequest]{
		Msg: &database.CreateCredentialsRequest{
			DatabaseId: id,
			ClientId:   clientID,
		},
	})
	if err != nil {
		return err
	}

	cmd.Printf("created credential - clientID: %s, secret: %s\n", clientID, res.Msg.GetClientSecret())
	return nil
}
