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
			header := []string{"Environment", "Triggers", "Fixed Fields"}
			rows := [][]string{}
			for _, p := range pipeline.Phases {
				var triggers []string
				for _, t := range p.GetTriggers() {
					triggers = append(triggers, triggerToString(t))
				}

				fieldsString := ""
				if len(p.GetFieldPrefixes().GetPrefixes()) > 0 {
					fieldsString = "Excluded: "
					if p.GetFieldPrefixes().GetInclusion() {
						fieldsString = "Included: "
					}

					fieldsString += strings.Join(p.GetFieldPrefixes().GetPrefixes(), "/n")
				}

				rows = append(rows, []string{p.GetEnvironmentId(), strings.Join(triggers, "\n"),
					fieldsString})
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

func updateTriggers(prompter Prompter, triggers []*model.PromotionTrigger) ([]*model.PromotionTrigger, error) {
	if len(triggers) == 0 {
		fmt.Println("No triggers configured - Let's configure one!")
		t, err := updateTrigger(prompter, nil)
		if err != nil {
			return nil, err
		}

		return []*model.PromotionTrigger{
			t,
		}, nil
	}

	triggerToRow := func(t *model.PromotionTrigger) []string {
		details := "None"
		if t.GetCondition() != nil {
			switch v := t.GetCondition().(type) {
			case *model.PromotionTrigger_TimeAlive:
				details = fmt.Sprintf("Time Alive: %s", v.TimeAlive.AsDuration().String())
			default:
				details = "Unknown"
			}
		}

		triggerType := "Manual"
		if t.GetAutomatic() {
			triggerType = "Auto"
		}

		return []string{triggerType, details}
	}

	header := []string{"Type", "Details"}
	var triggerRows [][]string
	for _, t := range triggers {
		triggerRows = append(triggerRows, triggerToRow(t))
	}
	triggerRows = append(triggerRows, []string{"Add new trigger", ""})
	triggerRows = append(triggerRows, []string{"Remove trigger", ""})
	triggerRows = append(triggerRows, []string{"Done", ""})

	for {
		i, err := prompter.TableSelect("Select the trigger to update (CTRL + C to cancel)", triggerRows, header)
		if err != nil {
			return nil, err
		}

		switch i {
		case len(triggerRows) - 1:
			return triggers, nil
		case len(triggerRows) - 2:
			// Remove trigger
			i, err := prompter.TableSelect("Select the trigger to remove (CTRL + C to cancel)",
				triggerRows[:len(triggerRows)-3], header)
			if err != nil {
				if ErrIsAborted(err) {
					continue
				}
				return nil, err
			}

			triggerRows = append(triggerRows[:i], triggerRows[i+1:]...)
			triggers = append(triggers[:i], triggers[i+1:]...)
		case len(triggerRows) - 3:
			// Add new trigger
			t, err := updateTrigger(prompter, nil)
			if err != nil {
				if ErrIsAborted(err) {
					continue
				}
				return nil, err
			}

			triggers = append(triggers, t)
			triggerRows = slices.Insert(triggerRows, len(triggerRows)-3, triggerToRow(t))
		default:
			// Update trigger
			t, err := updateTrigger(prompter, triggers[i])
			if err != nil {
				if ErrIsAborted(err) {
					continue
				}
				return nil, err
			}

			triggers[i] = t
			triggerRows[i] = triggerToRow(t)
		}
	}
}

func updateTrigger(prompter Prompter, trigger *model.PromotionTrigger) (*model.PromotionTrigger, error) {
	if trigger == nil {
		trigger = &model.PromotionTrigger{}
	}
	trigger = proto.Clone(trigger).(*model.PromotionTrigger)

	types := []string{"Auto", "Manual"}
	i, _, err := prompter.Select("Select the trigger type", types)
	if err != nil {
		return nil, err
	}

	if i == 0 {
		trigger.Automatic = true
	} else {
		trigger.Automatic = false
	}

	hasCondition, err := prompter.Confirm("Is this trigger condition based?", false)
	if err != nil {
		return nil, err
	}

	if !hasCondition {
		trigger.Condition = nil
		return trigger, nil
	}

	d, err := prompter.Input("Enter the time alive", ValidateDurationOpt,
		InputDefaultOpt(trigger.GetTimeAlive().AsDuration().String()))
	if err != nil {
		return nil, err
	}

	dur, err := time.ParseDuration(d)
	if err != nil {
		return nil, err
	}

	trigger.Condition = &model.PromotionTrigger_TimeAlive{
		TimeAlive: durationpb.New(dur),
	}

	return trigger, nil
}

func triggerToString(t *model.PromotionTrigger) string {
	str := "manual"
	if t.GetAutomatic() {
		str = "auto"
	}

	if t.GetCondition() != nil {
		switch v := t.GetCondition().(type) {
		case *model.PromotionTrigger_TimeAlive:
			str = str + fmt.Sprintf(" (%s)", v.TimeAlive.AsDuration().String())
		}
	}

	return str
}
