package database

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
)

func Connect(ctx context.Context, cmd *cobra.Command, args []string, nc rig.Client) error {
	identifier := ""
	if len(args) > 0 {
		identifier = args[0]
	}
	_, id, err := common.GetDatabase(ctx, identifier, nc)
	if err != nil {
		return err
	}

	if clientID == "" {
		clientID, err = common.PromptGetInput("Client ID", common.ValidateNonEmptyOpt)
		if err != nil {
			return err
		}
	}

	if clientSecret == "" {
		clientSecret, err = common.PromptGetInput("Client Secret", common.ValidateNonEmptyOpt)
		if err != nil {
			return err
		}
	}

	endpointRes, err := nc.Database().GetEndpoint(ctx, &connect.Request[database.GetEndpointRequest]{
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

	cmd.Printf("Connect to mongo by running: \n mongosh \"%s\"\n", endpoint)
	return nil
}
