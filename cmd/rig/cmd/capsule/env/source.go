package env

import (
	"context"
	"strings"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
)

var _kinds = map[string]capsule.EnvironmentSource_Kind{
	"configmap": capsule.EnvironmentSource_KIND_CONFIG_MAP,
	"secret":    capsule.EnvironmentSource_KIND_SECRET,
}

func (c *Cmd) source(ctx context.Context, _ *cobra.Command, args []string) error {
	if len(args) != 2 {
		return errors.InvalidArgumentErrorf("expected kind and name arguments")
	}

	cs := &capsule.ContainerSettings{}

	r, err := capsule_cmd.GetCurrentRollout(ctx, c.Rig, c.Cfg)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	if r.GetConfig().GetContainerSettings() != nil {
		cs = r.GetConfig().GetContainerSettings()
	}

	kind, ok := _kinds[strings.ToLower(args[0])]
	if !ok {
		return errors.InvalidArgumentErrorf("invalid kind, must be Secret or ConfigMap, got '%s'", args[1])
	}

	name := args[1]

	if remove {
		for i, es := range cs.GetEnvironmentSources() {
			if es.GetKind() == kind && es.GetName() == name {
				cs.EnvironmentSources = append(cs.EnvironmentSources[:i], cs.EnvironmentSources[i+1:]...)
				break
			}
		}
	} else {
		cs.EnvironmentSources = append(cs.EnvironmentSources, &capsule.EnvironmentSource{
			Name: name,
			Kind: kind,
		})
	}

	req := &connect.Request[capsule.DeployRequest]{
		Msg: &capsule.DeployRequest{
			CapsuleId: capsule_cmd.CapsuleID,
			Changes: []*capsule.Change{
				{
					Field: &capsule.Change_ContainerSettings{
						ContainerSettings: cs,
					},
				},
			},
			ProjectId:     flags.GetProject(c.Cfg),
			EnvironmentId: flags.GetEnvironment(c.Cfg),
		},
	}

	// TODO: Make helper for this this!
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
