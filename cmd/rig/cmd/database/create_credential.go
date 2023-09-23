package database

import (
	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/database"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
)

func (c Cmd) createCredentials(cmd *cobra.Command, args []string) error {
	ctx := c.Ctx
	identifier := ""
	if len(args) > 0 {
		identifier = args[0]
	}
	_, id, err := common.GetDatabase(ctx, identifier, c.Rig)
	if err != nil {
		return err
	}

	if name == "" {
		name, err = common.PromptInput("Credential Name:", common.ValidateNonEmptyOpt)
		if err != nil {
			return err
		}
	}
	res, err := c.Rig.Database().CreateCredential(ctx, &connect.Request[database.CreateCredentialRequest]{
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
