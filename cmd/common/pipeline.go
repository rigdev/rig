package common

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/rigdev/rig-go-api/api/v1/environment"
	"github.com/rigdev/rig-go-api/model"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
)

// var conditionTypes = []string{
// 	"Time Alive",
// }

func PromptPipelines(
	prompter Prompter,
	pipelines []*model.Pipeline,
	envs []*environment.Environment,
) ([]*model.Pipeline, error) {
	var envIDs []string
	for _, e := range envs {
		envIDs = append(envIDs, e.GetEnvironmentId())
	}

	if len(pipelines) == 0 {
		fmt.Println("No pipelines configured - Let's configure one!")
		p, err := updatePipeline(prompter, nil, envIDs)
		if err != nil {
			return nil, err
		}

		return []*model.Pipeline{
			p,
		}, nil
	}

	header := []string{"Name", "Start Env", "Phases"}
	rows := [][]string{}
	for _, p := range pipelines {
		var phases []string
		for _, phase := range p.GetPhases() {
			phases = append(phases, phase.GetEnvironmentId())
		}
		rows = append(rows, []string{p.GetName(), p.GetInitialEnvironment(), strings.Join(phases, ", ")})
	}

	rows = append(rows, []string{"Add new pipeline", "", ""})
	rows = append(rows, []string{"Delete a pipeline", "", ""})
	rows = append(rows, []string{"Done", "", ""})

	for {
		i, err := prompter.TableSelect("Select the pipeline to update (CTRL + C to cancel)", rows, header)
		if err != nil {
			return nil, err
		}

		switch i {
		case len(rows) - 1:
			return pipelines, nil
		case len(rows) - 2:
			// Delete pipeline
			i, err := prompter.TableSelect("Select the pipeline to delete (CTRL + C to cancel)", rows[:len(rows)-3], header)
			if err != nil {
				if ErrIsAborted(err) {
					continue
				}
				return nil, err
			}

			rows = append(rows[:i], rows[i+1:]...)
			pipelines = append(pipelines[:i], pipelines[i+1:]...)
		case len(rows) - 3:
			// Add new pipeline
			p, err := updatePipeline(prompter, nil, envIDs)
			if err != nil {
				if ErrIsAborted(err) {
					continue
				}
				return nil, err
			}

			pipelines = append(pipelines, p)
			var phases []string
			for _, phase := range p.GetPhases() {
				phases = append(phases, phase.GetEnvironmentId())
			}
			rows = slices.Insert(rows, len(rows)-3, []string{p.GetName(), p.GetInitialEnvironment(), strings.Join(phases, ", ")})
		default:
			// Update pipeline
			p, err := updatePipeline(prompter, pipelines[i], envIDs)
			if err != nil {
				if ErrIsAborted(err) {
					continue
				}
				return nil, err
			}

			var phases []string
			for _, phase := range p.GetPhases() {
				phases = append(phases, phase.GetEnvironmentId())
			}
			rows[i] = []string{p.GetName(), p.GetInitialEnvironment(), strings.Join(phases, ", ")}
			pipelines[i] = p
		}
	}
}

func updatePipeline(
	prompter Prompter,
	pipeline *model.Pipeline,
	envIDs []string,
) (*model.Pipeline, error) {
	if pipeline == nil {
		name, err := prompter.Input("Enter the pipeline name", ValidateNonEmptyOpt)
		if err != nil {
			return nil, err
		}
		pipeline = &model.Pipeline{
			Name: name,
		}
	}
	pipeline = proto.Clone(pipeline).(*model.Pipeline)

	fields := []string{
		"Initial Environment",
	}

	for _, p := range pipeline.Phases {
		fields = append(fields, fmt.Sprintf("Phase %s", p.GetEnvironmentId()))
	}

	fields = append(fields, []string{"Add new phase", "Remove phase", "Done"}...)

	for {
		i, _, err := prompter.Select("Select the field to update (CTRL + C to cancel)", fields)
		if err != nil {
			return nil, err
		}

		switch i {
		case 0:
			_, env, err := prompter.Select("Select the initial environment", envIDs)
			if err != nil {
				if ErrIsAborted(err) {
					continue
				}
				return nil, err
			}
			pipeline.InitialEnvironment = env
		case len(fields) - 3:
			// Add new phase
			phase, err := updatePhase(prompter, nil, envIDs)
			if err != nil {
				if ErrIsAborted(err) {
					continue
				}
				return nil, err
			}

			pipeline.Phases = append(pipeline.Phases, phase)
			fields = slices.Insert(fields, len(fields)-3, fmt.Sprintf("Phase %s", phase.GetEnvironmentId()))
		case len(fields) - 2:
			header := []string{"Environment", "Manual Trigger", "Auto Trigger", "Field Prefixes"}
			rows := [][]string{}
			for _, p := range pipeline.Phases {
				fmt.Println(p.GetTriggers())
				manualTrigger := triggerToString(p.GetTriggers().GetManual())
				autoTrigger := triggerToString(p.GetTriggers().GetAutomatic())

				fieldsString := ""
				if len(p.GetFieldPrefixes().GetPrefixes()) > 0 {
					fieldsString = "Excluded: "
					if p.GetFieldPrefixes().GetInclusion() {
						fieldsString = "Included: "
					}

					fieldsString += strings.Join(p.GetFieldPrefixes().GetPrefixes(), "/n")
				}

				rows = append(rows, []string{
					p.GetEnvironmentId(), manualTrigger, autoTrigger, fieldsString,
				})
			}
			i, err := prompter.TableSelect("Select the phase to update (CTRL + C to cancel)", rows, header)
			if err != nil {
				if ErrIsAborted(err) {
					continue
				}
				return nil, err
			}

			pipeline.Phases = append(pipeline.Phases[:i], pipeline.Phases[i+1:]...)
			fieldsIdx := i + 1
			fields = append(fields[:fieldsIdx], fields[fieldsIdx+1:]...)
		case len(fields) - 1:
			return pipeline, nil
		default:
			// Update phase
			phase, err := updatePhase(prompter, pipeline.Phases[i-1], envIDs)
			if err != nil {
				if ErrIsAborted(err) {
					continue
				}
				return nil, err
			}

			pipeline.Phases[i-1] = phase
			fields[i] = fmt.Sprintf("Phase %s", phase.GetEnvironmentId())
		}
	}
}

func updatePhase(
	prompter Prompter,
	phase *model.Phase,
	envIDs []string,
) (*model.Phase, error) {
	if phase == nil {
		phase = &model.Phase{}
	}
	phase = proto.Clone(phase).(*model.Phase)
	fields := []string{
		"Environment",
		"Triggers",
		"Fixed Fields",
		"Done",
	}

	for {
		i, _, err := prompter.Select("Select the field to update (CTRL + C to cancel)", fields)
		if err != nil {
			return nil, err
		}

		switch i {
		case 0:
			_, env, err := prompter.Select("Select the environment", envIDs)
			if err != nil {
				if ErrIsAborted(err) {
					continue
				}
				return nil, err
			}
			phase.EnvironmentId = env
		case 1:
			triggers, err := updateTriggers(prompter, phase.GetTriggers())
			if err != nil {
				if ErrIsAborted(err) {
					continue
				}
				return nil, err
			}

			phase.Triggers = triggers
		case 2:
			fieldPrefixes, err := updateFieldPrefixes(prompter, phase.GetFieldPrefixes())
			if err != nil {
				if ErrIsAborted(err) {
					continue
				}
				return nil, err
			}

			phase.FieldPrefixes = fieldPrefixes
		case 3:
			return phase, nil
		}
	}
}

func updateFieldPrefixes(prompter Prompter, prefixes *model.FieldPrefixes) (*model.FieldPrefixes, error) {
	if prefixes == nil {
		prefixes = &model.FieldPrefixes{}
	}
	prefixes = proto.Clone(prefixes).(*model.FieldPrefixes)
	inclusion := "Include Fields"
	if prefixes.Inclusion {
		inclusion = "Exclude Fields"
	}

	fields := append(prefixes.GetPrefixes(), inclusion, "Add", "Remove", "Done")
	for {
		i, _, err := prompter.Select("Select the field to update (CTRL + C to cancel)", fields)
		if err != nil {
			return nil, err
		}

		switch i {
		case len(fields) - 4:
			prefixes.Inclusion = !prefixes.Inclusion
			inclusion := "Include Fields"
			if prefixes.Inclusion {
				inclusion = "Exclude Fields"
			}

			fields[len(fields)-4] = inclusion
		case len(fields) - 3:
			field, err := prompter.Input("Enter the fixed field", ValidateNonEmptyOpt)
			if err != nil {
				if ErrIsAborted(err) {
					continue
				}
				return nil, err
			}

			prefixes.Prefixes = append(prefixes.GetPrefixes(), field)
			fields = append(fields[:len(fields)-4], field, "Add", "Remove", "Done")
		case len(fields) - 2:
			i, _, err := prompter.Select("Select the field to remove", prefixes.GetPrefixes())
			if err != nil {
				if ErrIsAborted(err) {
					continue
				}
				return nil, err
			}

			prefixes.Prefixes = append(prefixes.GetPrefixes()[:i], prefixes.GetPrefixes()[i+1:]...)
			fields = append(fields[:len(fields)-4], "Add", "Remove", "Done")
		case len(fields) - 1:
			return prefixes, nil
		default:
			field, err := prompter.Input("Enter the fixed field", ValidateNonEmptyOpt,
				InputDefaultOpt(prefixes.GetPrefixes()[i]))
			if err != nil {
				if ErrIsAborted(err) {
					continue
				}
				return nil, err
			}

			prefixes.Prefixes[i] = field
		}
	}
}

func updateTriggers(prompter Prompter, triggers *model.Triggers) (*model.Triggers, error) {
	if triggers == nil {
		triggers = &model.Triggers{}
	}
	triggers = proto.Clone(triggers).(*model.Triggers)

	triggerLabels := []string{
		"Auto " + triggerToString(triggers.GetAutomatic()),
		"Manual " + triggerToString(triggers.GetManual()),
		"Done",
	}

	for {
		i, _, err := prompter.Select("Select the trigger to update (CTRL + C to cancel)", triggerLabels)
		if err != nil {
			return nil, err
		}

		switch i {
		case 2:
			return triggers, nil
		case 1:
			t, err := updateTrigger(prompter, triggers.GetManual())
			if err != nil {
				if ErrIsAborted(err) {
					continue
				}
				return nil, err
			}

			triggers.Manual = t
			triggerLabels[1] = "Manual " + triggerToString(t)
		case 0:
			// Add new trigger
			t, err := updateTrigger(prompter, triggers.GetAutomatic())
			if err != nil {
				if ErrIsAborted(err) {
					continue
				}
				return nil, err
			}

			triggers.Automatic = t
			triggerLabels[0] = "Auto " + triggerToString(t)
		}
	}
}

func updateTrigger(prompter Prompter, trigger *model.Trigger) (*model.Trigger, error) {
	if trigger == nil {
		trigger = &model.Trigger{}
	}
	trigger = proto.Clone(trigger).(*model.Trigger)

	require := "Require all conditions"
	if trigger.GetRequireAll() {
		require = "Require any condition"
	}

	enable := "Enable"
	if trigger.GetEnabled() {
		enable = "Disable"
	}

	fields := []string{
		"Conditions",
		require,
		enable,
		"Clear",
		"Done",
	}

	for {
		i, _, err := prompter.Select("Select the field to update (CTRL + C to cancel)", fields)
		if err != nil {
			return nil, err
		}

		switch i {
		case 0:
			conditions, err := updateConditions(prompter, trigger.GetConditions())
			if err != nil {
				if ErrIsAborted(err) {
					continue
				}
				return nil, err
			}

			trigger.Conditions = conditions
		case 1:
			trigger.RequireAll = !trigger.RequireAll
			fields[0] = "Require all conditions"
			if trigger.GetRequireAll() {
				fields[0] = "Require any condition"
			}
		case 2:
			trigger.Enabled = !trigger.Enabled
			fields[1] = "Enable"
			if trigger.GetEnabled() {
				fields[1] = "Disable"
			}
		case 3:
			return nil, nil
		case 4:
			return trigger, nil
		}
	}
}

func updateConditions(prompter Prompter, conditions []*model.Trigger_Condition) ([]*model.Trigger_Condition, error) {
	conditionToLabel := func(c *model.Trigger_Condition) string {
		switch v := c.GetCondition().(type) {
		case *model.Trigger_Condition_TimeAlive:
			return fmt.Sprintf("Time Alive: %s", v.TimeAlive.AsDuration().String())
		}

		return "None"
	}

	conditionLabels := []string{}
	for _, c := range conditions {
		conditionLabels = append(conditionLabels, conditionToLabel(c))
	}
	conditionLabels = append(conditionLabels, "Add", "Remove", "Done")

	for {
		i, _, err := prompter.Select("Select the condition to update (CTRL + C to cancel)", conditionLabels)
		if err != nil {
			return nil, err
		}

		switch i {
		case len(conditionLabels) - 3:
			// When we have more, we should switch on types
			c, err := updateTimeAliveCondition(prompter, nil)
			if err != nil {
				if ErrIsAborted(err) {
					continue
				}
				return nil, err
			}

			conditions = append(conditions, c)

			conditionLabels = slices.Insert(conditionLabels,
				len(conditionLabels)-3, conditionToLabel(conditions[len(conditions)-1]))
		case len(conditionLabels) - 2:
			i, _, err := prompter.Select("Select the condition to remove", conditionLabels[:len(conditionLabels)-3])
			if err != nil {
				if ErrIsAborted(err) {
					continue
				}
				return nil, err
			}

			conditions = append(conditions[:i], conditions[i+1:]...)
			conditionLabels = append(conditionLabels[:i], conditionLabels[i+1:]...)
		case len(conditionLabels) - 1:
			return conditions, nil
		default:
			c, err := updateTimeAliveCondition(prompter, conditions[i])
			if err != nil {
				if ErrIsAborted(err) {
					continue
				}
				return nil, err
			}

			conditions[i] = c
			conditionLabels[i] = conditionToLabel(c)
		}
	}
}

func updateTimeAliveCondition(prompter Prompter, c *model.Trigger_Condition) (*model.Trigger_Condition, error) {
	if c == nil {
		c = &model.Trigger_Condition{}
	}

	c = proto.Clone(c).(*model.Trigger_Condition)
	d, err := prompter.Input("Enter the time alive", ValidateDurationOpt,
		InputDefaultOpt(c.GetTimeAlive().AsDuration().String()))
	if err != nil {
		return nil, err
	}

	dur, err := time.ParseDuration(d)
	if err != nil {
		return nil, err
	}

	c.Condition = &model.Trigger_Condition_TimeAlive{
		TimeAlive: durationpb.New(dur),
	}

	return c, nil
}

func triggerToString(t *model.Trigger) string {
	if t == nil {
		return "None"
	}

	if !t.GetEnabled() {
		return "Disabled"
	}

	conditions := []string{}
	for _, c := range t.GetConditions() {
		switch v := c.GetCondition().(type) {
		case *model.Trigger_Condition_TimeAlive:
			conditions = append(conditions, fmt.Sprintf("Time Alive (%s)", v.TimeAlive.AsDuration().String()))
		}
	}

	if len(conditions) == 0 {
		return "Instant"
	}

	if len(conditions) == 1 {
		return conditions[0]
	}

	require := "one-of"
	if t.GetRequireAll() {
		require = "all-of"
	}

	return fmt.Sprintf("%s: %s)", require, strings.Join(conditions, ", "))
}
