package capsule

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/fatih/color"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/utils"
)

var CapsuleID string

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

func formatCapsule(c *capsule.Capsule) string {
	var age string
	if c.GetCurrentRollout() == 0 {
		age = "-"
	} else {
		age = time.Since(c.GetUpdatedAt().AsTime()).Truncate(time.Second).String()
	}

	return fmt.Sprintf("%v\t (Rollout: %v, Updated At: %v)", c.GetCapsuleId(), c.GetCurrentRollout(), age)
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

func PromptAbortAndDeploy(ctx context.Context, capsuleID string, rig rig.Client, req *connect.Request[capsule.DeployRequest]) (*connect.Response[capsule.DeployResponse], error) {
	deploy, err := common.PromptConfirm("Rollout already in progress, would you like to cancel it and redeploy?", false)
	if err != nil {
		return nil, err
	}

	if !deploy {
		return nil, errors.FailedPreconditionErrorf("rollout already in progress")
	}

	return AbortAndDeploy(ctx, rig, capsuleID, req)
}

func AbortAndDeploy(ctx context.Context, rig rig.Client, capsuleID string, req *connect.Request[capsule.DeployRequest]) (*connect.Response[capsule.DeployResponse], error) {
	cc, err := rig.Capsule().Get(ctx, &connect.Request[capsule.GetRequest]{
		Msg: &capsule.GetRequest{
			CapsuleId: capsuleID,
		},
	})
	if err != nil {
		return nil, err
	}

	if _, err := rig.Capsule().AbortRollout(ctx, &connect.Request[capsule.AbortRolloutRequest]{
		Msg: &capsule.AbortRolloutRequest{
			CapsuleId: capsuleID,
			RolloutId: cc.Msg.GetCapsule().GetCurrentRollout(),
		},
	}); err != nil {
		return nil, err
	}

	return rig.Capsule().Deploy(ctx, req)
}

func Deploy(ctx context.Context, rig rig.Client, capsuleID string, req *connect.Request[capsule.DeployRequest], forceDeploy bool) error {
	_, err := rig.Capsule().Deploy(ctx, req)
	if errors.IsFailedPrecondition(err) && errors.MessageOf(err) == "rollout already in progress" {
		if forceDeploy {
			_, err = AbortAndDeploy(ctx, rig, capsuleID, req)
		} else {
			_, err = PromptAbortAndDeploy(ctx, capsuleID, rig, req)
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
			printInstanceID(stream.Msg().GetLog().GetInstanceId(), os.Stdout)
			os.Stdout.WriteString(stream.Msg().GetLog().GetTimestamp().AsTime().Format(base.RFC3339NanoFixed))
			os.Stdout.WriteString(": ")
			if _, err := os.Stdout.Write(v.Stdout); err != nil {
				return err
			}
		case *capsule.LogMessage_Stderr:
			printInstanceID(stream.Msg().GetLog().GetInstanceId(), os.Stderr)
			os.Stderr.WriteString(stream.Msg().GetLog().GetTimestamp().AsTime().Format(base.RFC3339NanoFixed))
			os.Stderr.WriteString(": ")
			if _, err := os.Stderr.Write(v.Stderr); err != nil {
				return err
			}
		default:
			return errors.InvalidArgumentErrorf("invalid log message")
		}
	}

	return stream.Err()
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

func printInstanceID(instanceID string, out *os.File) {
	c, ok := instanceToColor[instanceID]
	if !ok {
		c = colors[len(instanceToColor)%len(colors)]
		instanceToColor[instanceID] = c
	}
	color.Set(c)
	out.WriteString(instanceID + " ")
	color.Unset()
}
