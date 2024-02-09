package mount

import (
	"context"
	"fmt"
	"slices"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/cmd/common"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
)

func (c *Cmd) remove(ctx context.Context, cmd *cobra.Command, args []string) error {
	var path string
	var err error
	if len(args) != 1 {
		rollout, err := capsule_cmd.GetCurrentRollout(ctx, c.Rig, c.Cfg)
		if errors.IsNotFound(err) {
			cmd.Println("No config files mounted")
			return nil
		} else if err != nil {
			return err
		}

		var paths []string
		for _, path := range rollout.GetConfig().GetConfigFiles() {
			paths = append(paths, path.GetPath())
		}

		slices.Sort(paths)
		if len(paths) == 0 {
			fmt.Println("Capsule has no mounted files")
			return nil
		}

		_, path, err = common.PromptSelect("Mount path:", paths)
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
			ProjectId:     flags.GetProject(c.Cfg),
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

	cmd.Println(fmt.Sprintf("Config file %s removed", path))

	return nil
}
