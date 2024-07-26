package root

import (
	"context"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/cmd/common"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/spf13/cobra"
)

func (c *Cmd) create(ctx context.Context, cmd *cobra.Command, args []string) error {
	var err error
	if len(args) == 0 {
		capsule_cmd.CapsuleID, err = c.Prompter.Input("Capsule name:", common.ValidateSystemNameOpt)
		if err != nil {
			return err
		}
	} else {
		capsule_cmd.CapsuleID = args[0]
	}

	res, err := c.Rig.Capsule().Create(ctx, &connect.Request[capsule.CreateRequest]{
		Msg: &capsule.CreateRequest{
			Name:      capsule_cmd.CapsuleID,
			ProjectId: c.Scope.GetCurrentContext().GetProject(),
		},
	})
	if err != nil {
		return err
	}

	cmd.Printf("Created new capsule '%v'\n", res.Msg.GetCapsuleId())
	return nil
}
