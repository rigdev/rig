package utils

import (
	"slices"

	rollout_api "github.com/rigdev/rig-go-api/api/v1/capsule/rollout"
	settings_api "github.com/rigdev/rig-go-api/api/v1/settings"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/pkg/errors"
	"google.golang.org/protobuf/proto"
)

func ApplySettingsUpdates(settings *settings_api.Settings, updates []*settings_api.Update) error {
	for _, update := range updates {
		switch v := update.Field.(type) {
		case *settings_api.Update_SetNotificationNotifiers_:
			settings.NotificationNotifiers = v.SetNotificationNotifiers.GetNotifiers()
		case *settings_api.Update_SetGitStore:
			if proto.Equal(&model.GitStore{}, v.SetGitStore) {
				v.SetGitStore = nil
			}
			settings.GitStore = v.SetGitStore
		case *settings_api.Update_SetPipelines_:
			settings.Pipelines = v.SetPipelines.GetPipelines()
		case *settings_api.Update_AddRolloutMetric:
			if err := ValidateRolloutMetric(v.AddRolloutMetric); err != nil {
				return errors.InvalidArgumentErrorf("invalid rollout metric: %w", err)
			}
			if settings.Metrics == nil {
				settings.Metrics = &settings_api.Metrics{}
			}
			if settings.Metrics.RolloutMetrics == nil {
				settings.Metrics.RolloutMetrics = &settings_api.RolloutMetrics{}
			}
			settings.Metrics.RolloutMetrics.Metrics = InsertInList(
				settings.Metrics.RolloutMetrics.Metrics,
				v.AddRolloutMetric,
				func(m1, m2 *rollout_api.Metric) bool {
					return m1.GetName() == m2.GetName()
				},
			)
		case *settings_api.Update_RemoveRolloutMetric_:
			if m := settings.GetMetrics(); m != nil {
				if r := m.GetRolloutMetrics(); r != nil {
					r.Metrics = RemoveFromList(r.Metrics, v.RemoveRolloutMetric.GetName(), func(m *rollout_api.Metric, k string) bool {
						return m.GetName() == k
					})
				}
			}
		default:
			return errors.InvalidArgumentErrorf("unknown update type: %T", v)
		}
	}

	return nil
}

// TODO Store as enums somewhere instead
var (
	stageNames  = []string{"configure", "resource_creation", "running"}
	stageStates = []string{"deploying", "running", "stopped"}
	stepTypes   = map[string][]string{
		"configure":         {"configure_capsule", "configure_file", "configure_env", "comit", "generic"},
		"resource_creation": {"create_resource", "generic"},
		"running":           {"instances", "generic"},
	}
	stepStates     = []string{"ongoing", "failed", "done"}
	durationSince  = []string{"start", "stage_enter", "stage_state_enter", "step_enter", "step_state_enter"}
	conditionTypes = []string{"first_time_in_condition", "any_time_in_condition"}
)

func ValidateRolloutMetric(metric *rollout_api.Metric) error {
	if err := ValidateSystemName(metric.GetName()); err != nil {
		return errors.InvalidArgumentErrorf("name '%s' was malformed: %w", metric.GetName(), err)
	}

	if name := metric.GetStageName(); name != "" {
		if !slices.Contains(stageNames, name) {
			return errors.InvalidArgumentErrorf("stage name '%s', if given, must be one of %s", name, stageNames)
		}
	}
	if state := metric.GetStageState(); state != nil {
		for _, s := range state.GetStates() {
			if !slices.Contains(stageStates, s) {
				return errors.InvalidArgumentErrorf("invalid stage state '%s'. The valid ones are %s", s, stageStates)
			}
		}
	}

	if t := metric.GetStepType(); t != "" {
		if metric.GetStageName() == "" {
			return errors.InvalidArgumentErrorf("cannot supply step type without stage name")
		}
		if !slices.Contains(stepTypes[metric.GetStageName()], t) {
			return errors.InvalidArgumentErrorf("invalid step type '%s' for stage %s. The valid are %s", t, metric.GetStageName(), stepTypes[metric.GetStageName()])
		}
	}
	if state := metric.GetStepState(); state != nil {
		for _, s := range state.GetStates() {
			if !slices.Contains(stepStates, s) {
				return errors.InvalidArgumentErrorf("invalid step state '%s'. The valid ones are %s", s, stepStates)
			}
		}
	}

	if !slices.Contains(durationSince, metric.GetDurationSince()) {
		return errors.InvalidArgumentErrorf("invalid duration since '%s'. Must be one of %s", metric.GetDurationSince(), durationSince)
	}

	if !slices.Contains(conditionTypes, metric.GetConditionType()) {
		return errors.InvalidArgumentErrorf("invalid condition type '%s'. Must be one of %s", metric.GetConditionType(), conditionTypes)
	}

	return nil
}
