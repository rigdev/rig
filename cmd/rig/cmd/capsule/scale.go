package capsule

import (
	"context"
	"strconv"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
)

func CapsuleScale(ctx context.Context, cmd *cobra.Command, args []string, capsuleID CapsuleID, nc rig.Client) error {
	var r uint64
	if replicas == -1 {
		rString, err := common.PromptInput("Enter Replica Count:", common.ValidateNonEmptyOpt)
		if err != nil {
			return err
		}
		r, err = strconv.ParseUint(rString, 10, 32)
		if err != nil {
			return errors.InvalidArgumentErrorf("invalid replica count; %v", err)
		}

	} else {
		r = uint64(replicas)
	}

	cgs := []*capsule.Change{{
		Field: &capsule.Change_Replicas{
			Replicas: uint32(r),
		},
	}}
	if _, err := nc.Capsule().Deploy(ctx, &connect.Request[capsule.DeployRequest]{
		Msg: &capsule.DeployRequest{
			CapsuleId: capsuleID.String(),
			Changes:   cgs,
		},
	}); err != nil {
		return err
	}

	cmd.Println("Capsule scaled to", r, "replicas")

	return nil
}
