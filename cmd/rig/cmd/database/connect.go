package database

import (
	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
)

func (c Cmd) connect(cmd *cobra.Command, args []string) error {
	ctx := c.Ctx
	identifier := ""
	if len(args) > 0 {
		identifier = args[0]
	}
	_, id, err := common.GetDatabase(ctx, identifier, c.Rig)
	if err != nil {
		return err
	}

	if clientID == "" {
		clientID, err = common.PromptInput("Client ID:", common.ValidateNonEmptyOpt)
		if err != nil {
			return err
		}
	}

	if clientSecret == "" {
		clientSecret, err = common.PromptInput("Client Secret:", common.ValidateNonEmptyOpt)
		if err != nil {
			return err
		}
	}

	endpointRes, err := c.Rig.Database().GetEndpoint(ctx, &connect.Request[database.GetEndpointRequest]{
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
