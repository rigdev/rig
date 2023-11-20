package serviceaccount

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/service_account"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
)

func (c *Cmd) create(ctx context.Context, cmd *cobra.Command, _ []string) error {
	var name string
	var err error

	if name == "" {
		name, err = common.PromptInput("Name:", common.ValidateNonEmptyOpt)
		if err != nil {
			return err
		}
	}

	resp, err := c.Rig.ServiceAccount().Create(ctx, &connect.Request[service_account.CreateRequest]{
		Msg: &service_account.CreateRequest{
			Name: name,
		},
	})
	if err != nil {
		return err
	}

	cmd.Print("Service Account created\n")
	cmd.Printf("ID: %s\n", resp.Msg.GetClientId())
	cmd.Printf("Secret: %s\n", resp.Msg.GetClientSecret())

	return nil
}
