package mount

import (
	"context"
	"os"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/cmd/common"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
)

func (c *Cmd) set(ctx context.Context, _ *cobra.Command, _ []string) error {
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

	var paths []string
	rollout, err := capsule_cmd.GetCurrentRollout(ctx, c.Rig, c.Cfg)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	for _, p := range rollout.GetConfig().GetConfigFiles() {
		paths = append(paths, p.GetPath())
	}
	if dstPath == "" {
		dstPath, err = common.PromptInput("Destination path", common.ValidateAbsPathOpt, common.ValidateUniqueOpt(paths))
		if err != nil {
			return err
		}
	}

	cf := &capsule.Change_SetConfigFile{
		SetConfigFile: &capsule.Change_ConfigFile{
			Path:     dstPath,
			Content:  bytes,
			IsSecret: secret,
		},
	}

	req := &connect.Request[capsule.DeployRequest]{
		Msg: &capsule.DeployRequest{
			CapsuleId: capsule_cmd.CapsuleID,
			Changes: []*capsule.Change{
				{
					Field: cf,
				},
			},
			ProjectId:     c.Cfg.GetProject(),
			EnvironmentId: flags.GetEnvironment(c.Cfg),
		},
	}

	_, err = c.Rig.Capsule().Deploy(ctx, req)
	if errors.IsFailedPrecondition(err) && errors.MessageOf(err) == "rollout already in progress" {
		if forceDeploy {
			_, err = capsule_cmd.AbortAndDeploy(ctx, c.Rig, c.Cfg, capsule_cmd.CapsuleID, req)
		} else {
			_, err = capsule_cmd.PromptAbortAndDeploy(ctx, capsule_cmd.CapsuleID, c.Rig, c.Cfg, req)
		}
	}
	if err != nil {
		return err
	}
	return nil
}
