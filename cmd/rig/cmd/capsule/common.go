package capsule

import (
	"context"
	"fmt"
	"os"
	"strings"

	"connectrpc.com/connect"
	"github.com/fatih/color"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/pkg/cli"
	"github.com/rigdev/rig/cmd/rig/cmd/cmdconfig"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/utils"
)

var CapsuleID string

func GetCurrentContainerResources(
	ctx context.Context,
	client rig.Client,
	cfg *cmdconfig.Config,
) (*capsule.ContainerSettings, uint32, error) {
	rollout, err := GetCurrentRollout(ctx, client, cfg)
	if err != nil {
		return nil, 0, err
	}
	container := rollout.GetConfig().GetContainerSettings()
	if container == nil {
		container = &capsule.ContainerSettings{}
	}
	if container.Resources == nil {
		container.Resources = &capsule.Resources{}
	}

	utils.FeedDefaultResources(container.Resources)

	return container, rollout.GetConfig().GetReplicas(), nil
}

func GetCurrentNetwork(ctx context.Context, client rig.Client, cfg *cmdconfig.Config) (*capsule.Network, error) {
	rollout, err := GetCurrentRollout(ctx, client, cfg)
	if err != nil {
		return nil, err
	}
	return rollout.GetConfig().GetNetwork(), nil
}

func GetCurrentRollout(ctx context.Context, client rig.Client, cfg *cmdconfig.Config) (*capsule.Rollout, error) {
	return GetCurrentRolloutOfCapsule(ctx, client, cfg, CapsuleID)
}

func GetCurrentRolloutOfCapsule(
	ctx context.Context,
	client rig.Client,
	cfg *cmdconfig.Config,
	capsuleID string,
) (*capsule.Rollout, error) {
	r, err := client.Capsule().ListRollouts(ctx, connect.NewRequest(&capsule.ListRolloutsRequest{
		CapsuleId: capsuleID,
		Pagination: &model.Pagination{
			Offset:     0,
			Limit:      1,
			Descending: true,
		},
		ProjectId:     flags.GetProject(cfg),
		EnvironmentId: flags.GetEnvironment(cfg),
	}))
	if err != nil {
		return nil, err
	}

	for _, r := range r.Msg.GetRollouts() {
		return r, nil
	}

	return nil, errors.NotFoundErrorf("no rollout for capsule")
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

func PromptAbortAndDeploy(
	ctx context.Context,
	rig rig.Client,
	req *connect.Request[capsule.DeployRequest],
) (*connect.Response[capsule.DeployResponse], error) {
	deploy, err := common.PromptConfirm("Rollout already in progress, would you like to cancel it and redeploy?", false)
	if err != nil {
		return nil, err
	}

	if !deploy {
		return nil, errors.FailedPreconditionErrorf("rollout already in progress")
	}

	return AbortAndDeploy(ctx, rig, req)
}

func AbortAndDeploy(
	ctx context.Context,
	rig rig.Client,
	req *connect.Request[capsule.DeployRequest],
) (*connect.Response[capsule.DeployResponse], error) {
	req.Msg.Force = true
	return rig.Capsule().Deploy(ctx, req)
}

func Deploy(
	ctx context.Context,
	rig rig.Client,
	req *connect.Request[capsule.DeployRequest],
	forceDeploy bool,
) error {
	_, err := rig.Capsule().Deploy(ctx, req)
	if errors.IsFailedPrecondition(err) && errors.MessageOf(err) == "rollout already in progress" {
		if forceDeploy {
			_, err = AbortAndDeploy(ctx, rig, req)
		} else {
			_, err = PromptAbortAndDeploy(ctx, rig, req)
		}
	}
	if err != nil {
		return err
	}
	return nil
}

func PrintLogs(stream *connect.ServerStreamForClient[capsule.LogsResponse]) error {
	for stream.Receive() {
		switch v := stream.Msg().GetLog().GetMessage().GetMessage().(type) {
		case *capsule.LogMessage_Stdout:
			if err := printInstanceID(stream.Msg().GetLog().GetInstanceId(), os.Stdout); err != nil {
				return err
			}
			os.Stdout.WriteString(stream.Msg().GetLog().GetTimestamp().AsTime().Format(cli.RFC3339NanoFixed))
			os.Stdout.WriteString(": ")
			if _, err := os.Stdout.Write(v.Stdout); err != nil {
				return err
			}
		case *capsule.LogMessage_Stderr:
			if err := printInstanceID(stream.Msg().GetLog().GetInstanceId(), os.Stderr); err != nil {
				return err
			}
			os.Stderr.WriteString(stream.Msg().GetLog().GetTimestamp().AsTime().Format(cli.RFC3339NanoFixed))
			os.Stderr.WriteString(": ")
			if _, err := os.Stderr.Write(v.Stderr); err != nil {
				return err
			}
		case *capsule.LogMessage_ContainerTermination_:
			if err := printInstanceID(stream.Msg().GetLog().GetInstanceId(), os.Stderr); err != nil {
				return err
			}
			os.Stdout.WriteString(stream.Msg().GetLog().GetTimestamp().AsTime().Format(cli.RFC3339NanoFixed))
			os.Stdout.WriteString(" Container Terminated.\n\n")
		default:
			return errors.InvalidArgumentErrorf("invalid log message")
		}
	}

	return stream.Err()
}

func SelectCapsule(ctx context.Context, rc rig.Client, cfg *cmdconfig.Config) (string, error) {
	resp, err := rc.Capsule().List(ctx, connect.NewRequest(&capsule.ListRequest{
		Pagination: &model.Pagination{},
		ProjectId:  flags.GetProject(cfg),
	}))
	if err != nil {
		return "", err
	}

	var capsuleNames []string
	for _, c := range resp.Msg.GetCapsules() {
		capsuleNames = append(capsuleNames, c.GetCapsuleId())
	}

	if len(capsuleNames) == 0 {
		return "", errors.New("This project has no capsules. Create one, to get started")
	}

	_, name, err := common.PromptSelect("Capsule: ", capsuleNames, common.SelectFuzzyFilterOpt)
	if err != nil {
		return "", err
	}

	return name, nil
}

var colors = []color.Attribute{
	color.FgRed,
	color.FgBlue,
	color.FgCyan,
	color.FgGreen,
	color.FgYellow,
	color.FgMagenta,
	color.FgWhite,
}

var instanceToColor = map[string]color.Attribute{}

func printInstanceID(instanceID string, out *os.File) error {
	c, ok := instanceToColor[instanceID]
	if !ok {
		c = colors[len(instanceToColor)%len(colors)]
		instanceToColor[instanceID] = c
	}
	color.Set(c)
	if _, err := out.WriteString(instanceID + " "); err != nil {
		return fmt.Errorf("could not print instance id: %w", err)
	}
	color.Unset()
	return nil
}
