package database

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/rig/cmd/utils"
	"github.com/spf13/cobra"
)

func GetEndpoint(ctx context.Context, cmd *cobra.Command, args []string, rc rig.Client) error {
	identifier := ""
	if len(args) > 0 {
		identifier = args[0]
	}
	db, id, err := utils.GetDatabase(ctx, identifier, rc)
	if err != nil {
		return err
	}

	fmt.Println("creds", db.GetClientIds())

	if clientID == "" {
		clientID, err = utils.PromptGetInput("Client ID", utils.ValidateNonEmpty)
		if err != nil {
			return err
		}
	}

	if clientSecret == "" {
		clientSecret, err = utils.PromptGetInput("Client Secret", utils.ValidateNonEmpty)
		if err != nil {
			return err
		}
	}

	endpointRes, err := rc.Database().GetEndpoint(ctx, &connect.Request[database.GetEndpointRequest]{
		Msg: &database.GetEndpointRequest{
			DatabaseId:   id,
			ClientId:     clientID,
			ClientSecret: clientSecret,
		},
	})
	if err != nil {
		return err
	}

	endpoint := endpointRes.Msg.GetEndpoint()

	cmd.Printf("Connect to the database by running: \n mongosh \"%s\"\n", endpoint)
	return nil
}
