package mount

import (
	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/cmd/common"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
)

func (c Cmd) remove(cmd *cobra.Command, args []string) error {
	ctx := c.Ctx
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
	req := &connect.Request[capsule.DeployRequest]{
		Msg: &capsule.DeployRequest{
			CapsuleId: capsule_cmd.CapsuleID,
			Changes: []*capsule.Change{
				{
					Field: cf,
				},
			},
		},
	}

	_, err = c.Rig.Capsule().Deploy(ctx, req)
	if errors.IsFailedPrecondition(err) && errors.MessageOf(err) == "rollout already in progress" {
		_, err = capsule_cmd.PromptAbortAndDeploy(ctx, capsule_cmd.CapsuleID, c.Rig, req)
	}
	if err != nil {
		return err
	}

	return nil
}
