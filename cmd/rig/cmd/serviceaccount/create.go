package serviceaccount

import (
	"context"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/service_account"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
)

func (c *Cmd) create(ctx context.Context, cmd *cobra.Command, args []string) error {
	var name string
	var err error
	if len(args) < 1 {
		name, err = c.Prompter.Input("Name:", common.ValidateNonEmptyOpt)
		if err != nil {
			return err
		}
	} else {
		name = args[0]
	}

	if role == "" {
		_, role, err = c.Prompter.Select("What is the role of the user?",
			[]string{"admin", "owner", "developer", "viewer"})
		if err != nil {
			return err
		}
	}

	resp, err := c.Rig.ServiceAccount().Create(ctx, &connect.Request[service_account.CreateRequest]{
		Msg: &service_account.CreateRequest{
			Name:           name,
			InitialGroupId: role,
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
