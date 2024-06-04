package root

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	"github.com/fatih/color"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/operator/api/v1/pipeline"
	"github.com/rigdev/rig/cmd/common"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/spf13/cobra"
)

var boldWhite = color.New(color.FgWhite, color.Bold)
var red = color.New(color.FgRed)

func (c *Cmd) status(ctx context.Context, _ *cobra.Command, _ []string) error {
	statusResp, err := c.Rig.Capsule().GetStatus(ctx, &connect.Request[capsule.GetStatusRequest]{
		Msg: &capsule.GetStatusRequest{
			CapsuleId:     capsule_cmd.CapsuleID,
			ProjectId:     flags.GetProject(c.Scope),
			EnvironmentId: flags.GetEnvironment(c.Scope),
		},
	})
	if err != nil {
		return err
	}

	status := statusResp.Msg.GetStatus()
	if flags.Flags.OutputType != common.OutputTypePretty {
		return common.FormatPrint(status, flags.Flags.OutputType)
	}

	// rollout, err := c.Rig.Capsule().GetRollout(ctx, connect.NewRequest()&capsule.GetRolloutRequest{})

	printStatusSummary(status, capsule_cmd.CapsuleID)

	return nil
}
func printStatusSummary(s *capsule.Status, capsuleID string) {
	builder := &strings.Builder{}
	buildCapsuleInfo(builder, s, capsuleID)
	// buildRolloutStatus(builder, s.GetRollout())
	buildContainerConfig(builder, s.GetContainerConfig())
	buildInstanceStatus(builder, s.GetInstances())
	buildConfigFileStatus(builder, s.GetConfigFiles())
	buildInterfaceStatus(builder, s.GetInterfaces())
	buildCronjobStatus(builder, s.GetCronJobs())
	fmt.Println(builder.String())
}

func buildCapsuleInfo(builder *strings.Builder, s *capsule.Status, capsuleID string) {
	builder.WriteString(boldWhite.Sprintf("Capsule %s\n", capsuleID))
	builder.WriteString(getIndented(fmt.Sprintf("Namespace: %s", s.GetNamespace()), 2))
}

// func buildRolloutStatus(builder *strings.Builder, r *capsule.RolloutStatus) {
// 	builder.WriteString(boldWhite.Sprintf("Rollout %d\n", r.GetRolloutId()))
// 	createdAt := r.GetCreatedAt().AsTime().Format("2006-01-02 15:04:05")

// 	builder.WriteString(getIndented(fmt.Sprintf("Stage: %s", rolloutStageToString(r.GetCurrentStage())), 2))
// 	if r.GetCommitHash() != "" {
// 		builder.WriteString(getIndented(fmt.Sprintf("Commit: %s", r.GetCommitHash()), 2))
// 	}
// 	if r.GetCommitUrl() != "" {
// 		builder.WriteString(getIndented(fmt.Sprintf("Commit URL: %s", r.GetCommitUrl()), 2))
// 	}
// 	if r.GetCreatedBy() != nil {
// 		builder.WriteString(getIndented(fmt.Sprintf("Created by: %s", r.GetCreatedBy().GetPrintableName()), 2))
// 	}
// 	if !r.GetCreatedAt().AsTime().IsZero() {
// 		builder.WriteString(getIndented(fmt.Sprintf("Created at: %s", createdAt), 2))
// 	}
// }

func buildContainerConfig(builder *strings.Builder, c *capsule.ContainerConfig) {
	builder.WriteString(boldWhite.Sprintf("Container Config\n"))
	builder.WriteString(getIndented(fmt.Sprintf("Image: %s", c.GetImage()), 2))
	if c.GetCommand() != "" {
		builder.WriteString(getIndented(fmt.Sprintf("%s %s", c.GetCommand(), strings.Join(c.GetArgs(), " ")), 2))
	}
	if len(c.GetEnvironmentVariables()) > 0 {
		builder.WriteString(getIndented("Environment Variables", 2))
		for key, value := range c.GetEnvironmentVariables() {
			builder.WriteString(getIndented(fmt.Sprintf("%s=%s", key, value), 4))
		}
	}

	if c.GetScale().GetCpuTarget() == nil {
		builder.WriteString(getIndented(fmt.Sprintf("#Replicas: %d", c.GetScale().GetMinReplicas()), 2))
	} else {
		builder.WriteString(getIndented(fmt.Sprintf("Auto-scaling: %d-%d",
			c.GetScale().GetMinReplicas(), c.GetScale().GetMaxReplicas()), 2))
	}
}

func buildInstanceStatus(builder *strings.Builder, i *capsule.InstancesStatus) {
	builder.WriteString(boldWhite.Sprintf("Instances\n"))
	builder.WriteString(getIndented(fmt.Sprintf("#Healthy: %d", i.GetNumReady()), 2))
	builder.WriteString(getIndented(fmt.Sprintf("#Upgrading: %d", i.GetNumUpgrading()), 2))
	builder.WriteString(getIndented(fmt.Sprintf("#Old Version: %d", i.GetNumWrongVersion()), 2))
	if i.GetNumStuck() > 0 {
		builder.WriteString(getIndented(red.Sprintf("#Failing: %d", i.GetNumStuck()), 2))
	} else {
		builder.WriteString(getIndented(fmt.Sprintf("#Failing: %d", i.GetNumStuck()), 2))
	}
}

func buildConfigFileStatus(builder *strings.Builder, c []*capsule.ConfigFileStatus) {
	if len(c) == 0 {
		builder.WriteString(boldWhite.Sprintf("No Config Files\n"))
		return
	}

	builder.WriteString(boldWhite.Sprintf("Config Files\n"))
	for _, cf := range c {
		transition := transitionToIcon(cf.GetTransition())
		state := stateToIcon(getAggregatedStatus(cf.GetStatus()))
		builder.WriteString(getIndented(fmt.Sprintf("%s %s%s", cf.GetPath(), transition, state), 2))
	}
}

func buildInterfaceStatus(builder *strings.Builder, c []*capsule.InterfaceStatus) {
	if len(c) == 0 {
		builder.WriteString(boldWhite.Sprintf("No Interfaces\n"))
		return
	}
	builder.WriteString(boldWhite.Sprintf("Interfaces\n"))
	for _, i := range c {
		transition := transitionToIcon(i.GetTransition())
		state := stateToIcon(getAggregatedStatus(i.GetStatus()))
		builder.WriteString(getIndented(fmt.Sprintf("%s:%d %s%s", i.GetName(), i.GetPort(), transition, state), 2))
		for _, r := range i.GetRoutes() {
			state := stateToIcon(getAggregatedStatus(r.GetStatus()))
			transition := transitionToIcon(r.GetTransition())
			builder.WriteString(getIndented(fmt.Sprintf("%s %s%s", r.GetRoute().GetHost(), transition, state), 4))
		}
	}
}

func buildCronjobStatus(builder *strings.Builder, cs []*capsule.CronJobStatus) {
	if len(cs) == 0 {
		builder.WriteString(boldWhite.Sprintf("No Cron Jobs\n"))
		return
	}
	builder.WriteString(boldWhite.Sprintf("Cron Jobs\n"))
	for _, c := range cs {
		transition := transitionToIcon(c.GetTransition())
		state := stateToIcon(c.GetLastExecution())
		builder.WriteString(getIndented(fmt.Sprintf("%s %s%s", c.GetJobName(), transition, state), 2))
	}
}

func getAggregatedStatus(statuses []*pipeline.ObjectStatus) pipeline.ObjectState {
	if len(statuses) == 0 {
		return pipeline.ObjectState_OBJECT_STATE_UNSPECIFIED
	}
	state := pipeline.ObjectState_OBJECT_STATE_HEALTHY
	for _, s := range statuses {
		for _, c := range s.GetInfo().GetConditions() {
			if c.State == pipeline.ObjectState_OBJECT_STATE_PENDING {
				state = pipeline.ObjectState_OBJECT_STATE_PENDING
			}
			if c.State == pipeline.ObjectState_OBJECT_STATE_ERROR {
				return pipeline.ObjectState_OBJECT_STATE_ERROR
			}
		}
	}
	return state
}

func stateToIcon(state pipeline.ObjectState) string {
	switch state {
	case pipeline.ObjectState_OBJECT_STATE_HEALTHY:
		return "✅"
	case pipeline.ObjectState_OBJECT_STATE_PENDING:
		return "⏳"
	case pipeline.ObjectState_OBJECT_STATE_ERROR:
		return "❌"
	default:
		return ""
	}
}

func transitionToIcon(transition capsule.Transition) string {
	switch transition {
	case capsule.Transition_TRANSITION_BEING_CREATED:
		return "🔼"
	case capsule.Transition_TRANSITION_BEING_DELETED:
		return "🔽"
	default:
		return ""
	}
}

func getIndented(s string, indent int) string {
	return fmt.Sprintf("%s- %s\n", strings.Repeat(" ", indent), s)
}

// func rolloutStageToString(state rollout.State) string {
// 	switch state {
// 	case rollout.State_STATE_PREPARING:
// 		return "Preparing"
// 	case rollout.State_STATE_CONFIGURE:
// 		return "Configuring"
// 	case rollout.State_STATE_RESOURCE_CREATION:
// 		return "Resource Creation"
// 	case rollout.State_STATE_RUNNING:
// 		return "Running"
// 	case rollout.State_STATE_STOPPED:
// 		return "Stopped"
// 	default:
// 		return "Unknown"
// 	}
// }
