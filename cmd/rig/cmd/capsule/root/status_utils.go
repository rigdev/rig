package root

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/gdamore/tcell/v2"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/api/v1/capsule/rollout"
	"github.com/rigdev/rig-go-api/operator/api/v1/pipeline"
	"google.golang.org/protobuf/reflect/protoreflect"
)

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

func StageToString(StageType string, step *rollout.StepInfo) string {
	stepInfoString := ""
	if step != nil {
		stepInfoString = StepInfoToString(step)
	}
	return fmt.Sprintf("%s: %s", StageType, stepInfoString)
}

func StepInfoToString(info *rollout.StepInfo) string {
	return fmt.Sprintf("%s %s %s %s", info.GetUpdatedAt().AsTime().Format("2006-01-02 15:04:05"),
		info.GetName(), StepStateToIcon(info.GetState()), info.GetMessage())
}

func StateToIcon(state pipeline.ObjectState) string {
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

func StepStateToIcon(state rollout.StepState) string {
	switch state {
	case rollout.StepState_STEP_STATE_DONE:
		return "✅"
	case rollout.StepState_STEP_STATE_FAILED:
		return "❌"
	case rollout.StepState_STEP_STATE_ONGOING:
		return "⏳"
	default:
		return ""
	}
}

func StageStateToIcon(state rollout.StageState) string {
	switch state {
	case rollout.StageState_STAGE_STATE_DEPLOYING:
		return "❌"
	case rollout.StageState_STAGE_STATE_STOPPED:
		return "✅"
	case rollout.StageState_STAGE_STATE_RUNNING:
		return "⏳"
	default:
		return ""
	}
}

func StageStateToColor(state rollout.StageState) tcell.Color {
	switch state {
	case rollout.StageState_STAGE_STATE_DEPLOYING:
		return tcell.ColorRed
	case rollout.StageState_STAGE_STATE_STOPPED:
		return tcell.ColorGreen
	case rollout.StageState_STAGE_STATE_RUNNING:
		return tcell.ColorYellow
	default:
		return tcell.ColorWhite
	}
}

func StateToString(state pipeline.ObjectState) string {
	switch state {
	case pipeline.ObjectState_OBJECT_STATE_HEALTHY:
		return "Healthy"
	case pipeline.ObjectState_OBJECT_STATE_PENDING:
		return "Pending"
	case pipeline.ObjectState_OBJECT_STATE_ERROR:
		return "Failing"
	default:
		return "Unknown"
	}
}

func StateToTCellColor(state pipeline.ObjectState) tcell.Color {
	switch state {
	case pipeline.ObjectState_OBJECT_STATE_HEALTHY:
		return tcell.ColorGreen
	case pipeline.ObjectState_OBJECT_STATE_PENDING:
		return tcell.ColorYellow
	case pipeline.ObjectState_OBJECT_STATE_ERROR:
		return tcell.ColorRed
	default:
		return tcell.ColorWhite
	}
}

func StateToFatihColor(state pipeline.ObjectState) *color.Color {
	switch state {
	case pipeline.ObjectState_OBJECT_STATE_HEALTHY:
		return green
	case pipeline.ObjectState_OBJECT_STATE_PENDING:
		return yellow
	case pipeline.ObjectState_OBJECT_STATE_ERROR:
		return red
	default:
		return color.New(color.FgWhite)
	}
}

func TransitionToIcon(transition capsule.Transition) string {
	switch transition {
	case capsule.Transition_TRANSITION_BEING_CREATED:
		return "🔼"
	case capsule.Transition_TRANSITION_BEING_DELETED:
		return "🔽"
	default:
		return ""
	}
}

func TransitionToString(transition capsule.Transition) string {
	switch transition {
	case capsule.Transition_TRANSITION_BEING_CREATED:
		return "Being Created"
	case capsule.Transition_TRANSITION_BEING_DELETED:
		return "Being Deleted"
	case capsule.Transition_TRANSITION_UP_TO_DATE:
		return "Up to Date"
	default:
		return "Unknown"
	}
}

func RolloutStageToString(state rollout.State) string {
	switch state {
	case rollout.State_STATE_PREPARING:
		return "Preparing"
	case rollout.State_STATE_CONFIGURE:
		return "Configuring"
	case rollout.State_STATE_RESOURCE_CREATION:
		return "Resource Creation"
	case rollout.State_STATE_RUNNING:
		return "Running"
	case rollout.State_STATE_STOPPED:
		return "Stopped"
	default:
		return "Unknown"
	}
}

func GetIndented(s string, indent int) string {
	return fmt.Sprintf("%s- %s\n", strings.Repeat(" ", indent), s)
}

func GenericObjectStatusToString(s *pipeline.ObjectStatus, builder *strings.Builder) {
	builder.WriteString(GetIndented("Object:", 2))
	builder.WriteString(GetIndented(fmt.Sprintf("Name: %s", s.GetObjectRef().GetName()), 4))
	builder.WriteString(GetIndented(fmt.Sprintf("Namespace: %s", s.GetObjectRef().GetNamespace()), 4))
	builder.WriteString(GetIndented(fmt.Sprintf("GVK: %s", GvkToString(s.GetObjectRef().GetGvk())), 4))

	if len(s.GetInfo().GetConditions()) == 0 {
		builder.WriteString(GetIndented("No conditions", 2))
	} else {
		builder.WriteString(GetIndented("Conditions:", 2))
		for _, c := range s.GetInfo().GetConditions() {
			builder.WriteString(GetIndented(fmt.Sprintf("Name: %s", c.GetName()), 4))
			builder.WriteString(GetIndented(StateToFatihColor(c.GetState()).
				Sprintf("State: %s", StateToString(c.GetState())), 6))
			builder.WriteString(GetIndented(fmt.Sprintf("Message: %s", c.GetMessage()), 6))
		}
	}

	if len(s.GetInfo().GetProperties()) == 0 {
		builder.WriteString(GetIndented("No properties", 2))
	} else {
		builder.WriteString(GetIndented("Properties:", 2))
		for key, value := range s.GetInfo().GetProperties() {
			builder.WriteString(GetIndented(fmt.Sprintf("%s: %s\n", key, value), 4))
		}
	}
}

func GvkToString(gvk *pipeline.GVK) string {
	return fmt.Sprintf("%s/%s/%s", gvk.GetGroup(), gvk.GetVersion(), gvk.GetKind())
}

func PathMatchToString(match capsule.PathMatchType) string {
	switch match {
	case capsule.PathMatchType_PATH_MATCH_TYPE_EXACT:
		return "Exact"
	case capsule.PathMatchType_PATH_MATCH_TYPE_PATH_PREFIX:
		return "Prefix"
	default:
		return "Unknown"
	}
}

type getInfoStep interface {
	GetInfo() *rollout.StepInfo
}

func GetStepInfo(s any) *rollout.StepInfo {
	msg, ok := s.(protoreflect.ProtoMessage)
	if !ok {
		return nil
	}
	var info *rollout.StepInfo
	msg.ProtoReflect().Range(func(_ protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		if getInfoStep, ok := v.Message().Interface().(getInfoStep); ok {
			info = getInfoStep.GetInfo()
			return false
		}
		return true
	})

	return info
}
