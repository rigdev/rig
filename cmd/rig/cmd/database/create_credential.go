package database

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
)

func CreateCredential(ctx context.Context, cmd *cobra.Command, args []string, nc rig.Client) error {
	identifier := ""
	if len(args) > 0 {
		identifier = args[0]
	}
	_, id, err := common.GetDatabase(ctx, identifier, nc)
	if err != nil {
		return err
	}

	if name == "" {
		name, err = common.PromptGetInput("Credential Name", common.ValidateNonEmpty)
		if err != nil {
			return err
		}
	}
	res, err := nc.Database().CreateCredential(ctx, &connect.Request[database.CreateCredentialRequest]{
		Msg: &database.CreateCredentialRequest{
			DatabaseId: id,
			Name:       name,
		},
	})
	if err != nil {
		return err
	}

	cmd.Printf("created credential - clientID: %s, secret: %s\n", res.Msg.GetClientId(), res.Msg.GetClientSecret())
	return nil
}
