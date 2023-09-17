package capsule

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/rigdev/rig/cmd/rig/cmd/cmd_config"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/utils"
	"github.com/spf13/cobra"
)

func GetCurrentContainerResources(ctx context.Context, client rig.Client) (*capsule.ContainerSettings, uint32, error) {
	resp, err := client.Capsule().Get(ctx, connect.NewRequest(&capsule.GetRequest{
		CapsuleId: CapsuleID,
	}))
	if err != nil {
		return nil, 0, err
	}

	r, err := client.Capsule().GetRollout(ctx, connect.NewRequest(&capsule.GetRolloutRequest{
		CapsuleId: CapsuleID,
		RolloutId: resp.Msg.GetCapsule().GetCurrentRollout(),
	}))
	if err != nil {
		return nil, 0, err
	}

	container := r.Msg.GetRollout().GetConfig().GetContainerSettings()
	if container == nil {
		container = &capsule.ContainerSettings{}
	}
	if container.Resources == nil {
		container.Resources = &capsule.Resources{}
	}

	utils.FeedDefaultResources(container.Resources)

	return container, r.Msg.GetRollout().GetConfig().GetReplicas(), nil
}

func GetCurrentNetwork(ctx context.Context, client rig.Client) (*capsule.Network, error) {
	resp, err := client.Capsule().Get(ctx, connect.NewRequest(&capsule.GetRequest{
		CapsuleId: CapsuleID,
	}))
	if err != nil {
		return nil, err
	}

	r, err := client.Capsule().GetRollout(ctx, connect.NewRequest(&capsule.GetRolloutRequest{
		CapsuleId: CapsuleID,
		RolloutId: resp.Msg.GetCapsule().GetCurrentRollout(),
	}))
	if err != nil {
		return nil, err
	}

	return r.Msg.GetRollout().GetConfig().GetNetwork(), nil
}

func GetCurrentRollout(ctx context.Context, client rig.Client) (*capsule.Rollout, error) {
	resp, err := client.Capsule().Get(ctx, connect.NewRequest(&capsule.GetRequest{
		CapsuleId: CapsuleID,
	}))
	if err != nil {
		return nil, err
	}

	r, err := client.Capsule().GetRollout(ctx, connect.NewRequest(&capsule.GetRolloutRequest{
		CapsuleId: CapsuleID,
		RolloutId: resp.Msg.GetCapsule().GetCurrentRollout(),
	}))
	if err != nil {
		return nil, err
	}

	return r.Msg.GetRollout(), nil
}

var capsuleCompletions = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	capsuleIDs := []string{}

	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{}
	}
	cmd.Annotations[base.OmitUser] = "true"

	f := base.Register(
		func(ctx context.Context, rc rig.Client, cfg *cmd_config.Config) error {
			if cfg.GetCurrentContext() == nil || cfg.GetCurrentAuth() == nil {
				return errors.UnauthenticatedErrorf("")
			}

			resp, err := rc.Capsule().List(ctx, &connect.Request[capsule.ListRequest]{
				Msg: &capsule.ListRequest{},
			})
			if err != nil {
				return err
			}

			for _, c := range resp.Msg.GetCapsules() {
				if strings.HasPrefix(c.GetCapsuleId(), toComplete) {
					capsuleIDs = append(capsuleIDs, formatCapsule(c))
				}
			}

			return nil
		},
	)

	if err := f(cmd, args); err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	cmd.Annotations[base.OmitUser] = ""

	if len(capsuleIDs) == 0 {
		return nil, cobra.ShellCompDirectiveError
	}

	return capsuleIDs, cobra.ShellCompDirectiveDefault
}

var BuildCompletions = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveError
	}

	if CapsuleID == "" {
		return nil, cobra.ShellCompDirectiveError
	}

	buildIds := []string{}

	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{}
	}
	cmd.Annotations[base.OmitUser] = "true"

	f := base.Register(
		func(ctx context.Context, rc rig.Client, cfg *cmd_config.Config) error {
			if cfg.GetCurrentContext() == nil || cfg.GetCurrentAuth() == nil {
				return errors.UnauthenticatedErrorf("")
			}

			resp, err := rc.Capsule().ListBuilds(ctx, &connect.Request[capsule.ListBuildsRequest]{
				Msg: &capsule.ListBuildsRequest{
					CapsuleId: CapsuleID,
				},
			})
			if err != nil {
				return err
			}

			for _, b := range resp.Msg.GetBuilds() {
				if strings.HasPrefix(b.GetBuildId(), toComplete) {
					buildIds = append(buildIds, formatBuild(b))
				}
			}

			return nil
		},
	)

	if err := f(cmd, args); err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	cmd.Annotations[base.OmitUser] = ""

	if len(buildIds) == 0 {
		return nil, cobra.ShellCompDirectiveError
	}

	return buildIds, cobra.ShellCompDirectiveDefault
}

var RolloutCompletions = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if CapsuleID == "" {
		return nil, cobra.ShellCompDirectiveError
	}

	rolloutIds := []string{}

	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{}
	}
	cmd.Annotations[base.OmitUser] = "true"

	f := base.Register(
		func(ctx context.Context, rc rig.Client, cfg *cmd_config.Config) error {
			if cfg.GetCurrentContext() == nil || cfg.GetCurrentAuth() == nil {
				return errors.UnauthenticatedErrorf("")
			}

			resp, err := rc.Capsule().ListRollouts(ctx, &connect.Request[capsule.ListRolloutsRequest]{
				Msg: &capsule.ListRolloutsRequest{
					CapsuleId: CapsuleID,
				},
			})
			if err != nil {
				return err
			}

			for _, r := range resp.Msg.GetRollouts() {
				if strings.HasPrefix(fmt.Sprint(r.GetRolloutId()), toComplete) {
					rolloutIds = append(rolloutIds, formatRollout(r))
				}
			}

			return nil
		},
	)

	if err := f(cmd, args); err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	cmd.Annotations[base.OmitUser] = ""

	if len(rolloutIds) == 0 {
		return nil, cobra.ShellCompDirectiveError
	}

	return rolloutIds, cobra.ShellCompDirectiveDefault
}

var InstanceCompletions = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if CapsuleID == "" {
		return nil, cobra.ShellCompDirectiveError
	}

	instanceIds := []string{}

	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{}
	}
	cmd.Annotations[base.OmitUser] = "true"

	f := base.Register(
		func(ctx context.Context, rc rig.Client, cfg *cmd_config.Config) error {
			if cfg.GetCurrentContext() == nil || cfg.GetCurrentAuth() == nil {
				return errors.UnauthenticatedErrorf("")
			}

			resp, err := rc.Capsule().ListInstances(ctx, &connect.Request[capsule.ListInstancesRequest]{
				Msg: &capsule.ListInstancesRequest{
					CapsuleId: CapsuleID,
				},
			})
			if err != nil {
				return err
			}

			for _, i := range resp.Msg.GetInstances() {
				if strings.HasPrefix(fmt.Sprint(i.GetInstanceId()), toComplete) {
					instanceIds = append(instanceIds, formatInstance(i))
				}
			}

			return nil
		},
	)

	if err := f(cmd, args); err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	cmd.Annotations[base.OmitUser] = ""

	if len(instanceIds) == 0 {
		return nil, cobra.ShellCompDirectiveError
	}

	return instanceIds, cobra.ShellCompDirectiveDefault
}

var NetworkCompletions = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if CapsuleID == "" {
		return nil, cobra.ShellCompDirectiveError
	}

	interfaces := []string{}

	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{}
	}
	cmd.Annotations[base.OmitUser] = "true"

	f := base.Register(
		func(ctx context.Context, rc rig.Client, cfg *cmd_config.Config) error {
			if cfg.GetCurrentContext() == nil || cfg.GetCurrentAuth() == nil {
				return errors.UnauthenticatedErrorf("")
			}

			n, err := GetCurrentNetwork(ctx, rc)
			if err != nil {
				return err
			}

			for _, i := range n.GetInterfaces() {
				if strings.HasPrefix(i.GetName(), toComplete) {
					interfaces = append(interfaces, i.GetName())
				}
			}

			return nil
		},
	)

	if err := f(cmd, args); err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	cmd.Annotations[base.OmitUser] = ""

	if len(interfaces) == 0 {
		return nil, cobra.ShellCompDirectiveError
	}

	return interfaces, cobra.ShellCompDirectiveDefault
}

var MountCompletions = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if CapsuleID == "" {
		return nil, cobra.ShellCompDirectiveError
	}

	paths := []string{}

	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{}
	}
	cmd.Annotations[base.OmitUser] = "true"

	f := base.Register(
		func(ctx context.Context, rc rig.Client, cfg *cmd_config.Config) error {
			if cfg.GetCurrentContext() == nil || cfg.GetCurrentAuth() == nil {
				return errors.UnauthenticatedErrorf("")
			}

			r, err := GetCurrentRollout(ctx, rc)
			if err != nil {
				return err
			}

			for _, f := range r.GetConfig().GetConfigFiles() {
				if strings.HasPrefix(f.GetPath(), toComplete) {
					paths = append(paths, formatMount(f))
				}
			}

			return nil
		},
	)

	if err := f(cmd, args); err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	cmd.Annotations[base.OmitUser] = ""

	if len(paths) == 0 {
		return nil, cobra.ShellCompDirectiveError
	}

	return paths, cobra.ShellCompDirectiveDefault
}

var EnvCompletions = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if CapsuleID == "" {
		return nil, cobra.ShellCompDirectiveError
	}

	envKeys := []string{}

	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{}
	}
	cmd.Annotations[base.OmitUser] = "true"

	f := base.Register(
		func(ctx context.Context, rc rig.Client, cfg *cmd_config.Config) error {
			if cfg.GetCurrentContext() == nil || cfg.GetCurrentAuth() == nil {
				return errors.UnauthenticatedErrorf("")
			}

			r, err := GetCurrentRollout(ctx, rc)
			if err != nil {
				return err
			}

			for k := range r.GetConfig().GetContainerSettings().GetEnvironmentVariables() {
				if strings.HasPrefix(k, toComplete) {
					envKeys = append(envKeys, k)
				}
			}

			return nil
		},
	)

	if err := f(cmd, args); err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	cmd.Annotations[base.OmitUser] = ""

	if len(envKeys) == 0 {
		return nil, cobra.ShellCompDirectiveError
	}

	return envKeys, cobra.ShellCompDirectiveDefault
}

func formatCapsule(c *capsule.Capsule) string {
	var age string
	if c.GetCurrentRollout() == 0 {
		age = "-"
	} else {
		age = time.Since(c.GetUpdatedAt().AsTime()).Truncate(time.Second).String()
	}

	return fmt.Sprintf("%v\t (Rollout: %v, Updated At: %v)", c.GetCapsuleId(), c.GetCurrentRollout(), age)
}

func formatRollout(r *capsule.Rollout) string {
	return fmt.Sprintf("%v\t (State: %v)", r.GetRolloutId(), r.GetStatus().GetState())
}

func formatInstance(i *capsule.Instance) string {
	var startedAt string
	if i.GetStartedAt().AsTime().IsZero() {
		startedAt = "-"
	} else {
		startedAt = time.Since(i.GetStartedAt().AsTime()).Truncate(time.Second).String()
	}

	return fmt.Sprintf("%v\t (State: %v, Started At: %v)", i.GetInstanceId(), i.GetState(), startedAt)
}

func formatBuild(b *capsule.Build) string {
	var age string
	if b.GetCreatedAt().AsTime().IsZero() {
		age = "-"
	} else {
		age = time.Since(b.GetCreatedAt().AsTime()).Truncate(time.Second).String()
	}

	return fmt.Sprintf("%v\t (Age: %v)", b.GetBuildId(), age)
}

func formatMount(m *capsule.ConfigFile) string {
	var age string
	if m.GetUpdatedAt().AsTime().IsZero() {
		age = "-"
	} else {
		age = time.Since(m.GetUpdatedAt().AsTime()).Truncate(time.Second).String()
	}

	return fmt.Sprintf("%v\t (Age: %v)", m.GetPath(), age)
}

func Truncated(str string, max int) string {
	if len(str) > max {
		return str[:strings.LastIndexAny(str[:max], " .,:;-")] + "..."
	}

	return str
}

func TruncatedFixed(str string, max int) string {
	if len(str) > max {
		return str[:max] + "..."
	}

	return str
}
