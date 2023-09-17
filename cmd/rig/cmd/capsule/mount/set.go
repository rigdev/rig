package mount

import (
	"context"
	"os"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/spf13/cobra"
)

func set(ctx context.Context, args []string, cmd *cobra.Command, rc rig.Client) error {
	var err error
	if srcPath == "" {
		srcPath, err = common.PromptInput("Source path", common.ValidateFilePathOpt)
		if err != nil {
			return err
		}
	}

	bytes, err := os.ReadFile(srcPath)
	if err != nil {
		return err
	}

	if dstPath == "" {
		dstPath, err = common.PromptInput("Destination path", common.ValidateAbsPathOpt)
		if err != nil {
			return err
		}
	}

	cf := &capsule.Change_SetConfigFile{
		SetConfigFile: &capsule.Change_ConfigFile{
			Path:    dstPath,
			Content: bytes,
		},
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
