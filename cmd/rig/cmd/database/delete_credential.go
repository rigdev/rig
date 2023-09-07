package database

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
)

func DeleteCredential(ctx context.Context, cmd *cobra.Command, args []string, nc rig.Client) error {
	identifier := ""
	if len(args) > 0 {
		identifier = args[0]
	}
	_, id, err := common.GetDatabase(ctx, identifier, nc)
	if err != nil {
		return err
	}

	if name == "" {
		name, err = common.PromptGetInput("Credential Name", common.ValidateNonEmptyOpt)
		if err != nil {
			return err
		}
	}

	if _, err := nc.Database().DeleteCredential(ctx, &connect.Request[database.DeleteCredentialRequest]{
		Msg: &database.DeleteCredentialRequest{
			DatabaseId:     id,
			CredentialName: name,
		},
	}); err != nil {
		return err
	}

	cmd.Printf("deleted credential %s\n", name)
	return nil
}
