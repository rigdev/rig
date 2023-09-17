package mount

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/spf13/cobra"
)

func remove(ctx context.Context, args []string, cmd *cobra.Command, rc rig.Client) error {
	var path string
	var err error
	if len(args) != 1 {
		path, err = common.PromptInput("mount path:", common.ValidateAbsPathOpt)
		if err != nil {
			return err
		}
	} else {
		path = args[0]
	}

	cf := &capsule.Change_RemoveConfigFile{
		RemoveConfigFile: path,
	}

	if _, err := rc.Capsule().Deploy(ctx, &connect.Request[capsule.DeployRequest]{
		Msg: &capsule.DeployRequest{
			CapsuleId: capsule_cmd.CapsuleID,
			Changes: []*capsule.Change{
				{
					Field: cf,
				},
			},
		},
	}); err != nil {
		return err
	}

	return nil
}
