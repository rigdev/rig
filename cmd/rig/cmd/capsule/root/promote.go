package root

import (
	"bytes"
	"context"
	"fmt"
	"slices"

	"connectrpc.com/connect"
	"github.com/gdamore/tcell/v2"
	"github.com/homeport/dyff/pkg/dyff"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/api/v1/environment"
	"github.com/rigdev/rig-go-api/model"
	platformv1 "github.com/rigdev/rig-go-api/platform/v1"
	"github.com/rigdev/rig/cmd/common"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/field"
	"github.com/rivo/tview"
	"github.com/spf13/cobra"
)

func (c *Cmd) promote(ctx context.Context, cmd *cobra.Command, args []string) error {
	capsuleID := capsule_cmd.CapsuleID
	fromEnvID := ""
	toEnvID := ""
	var err error
	if len(args) > 2 {
		fromEnvID = args[1]
		toEnvID = args[2]
	} else if len(args) > 1 {
		if !c.Scope.IsInteractive() {
			return errors.InvalidArgumentErrorf("from and to environment IDs must be provided")
		}

		fromEnvID = args[1]
	}

	if fromEnvID == "" {
		fromEnvID, err = c.promptForEnvironment(ctx)
		if err != nil {
			return err
		}
	}

	if toEnvID == "" {
		toEnvID, err = c.promptForEnvironment(ctx)
		if err != nil {
			return err
		}
	}

	if fromEnvID == toEnvID {
		return errors.InvalidArgumentErrorf("from and to environment IDs must be different")
	}

	fromRollout, err := c.getLatestRollout(ctx, capsuleID, fromEnvID, flags.GetProject(c.Scope))
	if err != nil {
		return err
	}
	fromSpec := fromRollout.GetSpec()

	toRollout, err := c.getLatestRollout(ctx, capsuleID, toEnvID, flags.GetProject(c.Scope))
	if err != nil {
		return err
	}
	toSpec := toRollout.GetSpec()

	changes, err := field.Compare(toSpec, fromSpec, field.SpecKeys...)
	if err != nil {
		return err
	}

	if len(changes.Changes) == 0 {
		cmd.Println("No changes detected between the two environments")
		return nil
	}

	if !force {
		if !c.Scope.IsInteractive() {
			common.FormatPrint(changes, common.OutputTypeYAML)
			return errors.InvalidArgumentErrorf("Changes detected, but not in interactive mode to validate and edit them." +
				"Run again with --force to force a promotion without validating the differences " +
				"and/or with --dry-run to see the result of the promotion without it taking effect.")
		}

		// Interactively prompt the changes
		changes, err = c.promptChanges(ctx, toRollout.Spec, changes)
		if err != nil {
			return err
		}
	}

	if len(changes.Changes) == 0 {
		cmd.Println("No changes left to promote")
		return nil
	}

	if !dryRun {
		promote, err := c.Prompter.Confirm("Are you sure you want to promote these changes?", false)
		if err != nil {
			return err
		}

		if !promote {
			return nil
		}
	}

	applied, err := field.Apply(toRollout.Spec, changes.Changes)
	if err != nil {
		return err
	}

	change := &capsule.Change{
		Field: &capsule.Change_Spec{
			Spec: applied.(*platformv1.CapsuleSpec),
		},
	}

	baseInput := capsule_cmd.BaseInput{
		Ctx:           ctx,
		Rig:           c.Rig,
		CapsuleID:     capsuleID,
		EnvironmentID: toEnvID,
		ProjectID:     flags.GetProject(c.Scope),
	}

	if dryRun {
		input := capsule_cmd.DeployDryInput{
			BaseInput:        baseInput,
			Changes:          []*capsule.Change{change},
			Scheme:           c.Scheme,
			CurrentRolloutID: toRollout.GetRolloutId(),
			IsInteractive:    c.Scope.IsInteractive(),
		}

		return capsule_cmd.DeployDry(input)
	}

	rollbackID := toRollout.GetRolloutId()
	if noRollback {
		rollbackID = 0
	}

	input := capsule_cmd.DeployAndWaitInput{
		DeployInput: capsule_cmd.DeployInput{
			BaseInput:        baseInput,
			Changes:          []*capsule.Change{change},
			ForceDeploy:      true,
			CurrentRolloutID: toRollout.GetRolloutId(),
		},
		Timeout:    timeout,
		RollbackID: rollbackID,
		NoWait:     noWait,
	}

	return capsule_cmd.DeployAndWait(input)
}

func (c *Cmd) promptForEnvironment(ctx context.Context) (string, error) {
	res, err := c.Rig.Environment().List(ctx, connect.NewRequest(&environment.ListRequest{}))
	if err != nil {
		return "", err
	}

	var es []string
	for _, e := range res.Msg.GetEnvironments() {
		es = append(es, e.GetEnvironmentId())
	}

	for {
		i, _, err := c.Prompter.Select("Environment: ", es)
		if err != nil {
			return "", err
		}

		environment := res.Msg.GetEnvironments()[i]
		if flags.GetProject(c.Scope) != "" && !environment.GetGlobal() &&
			!slices.Contains(environment.GetActiveProjects(), flags.GetProject(c.Scope)) {
			selectNew, err := c.Prompter.Confirm(
				fmt.Sprintf(
					"Warning: project '%s' is not active in environment '%s'.\nDo you want to select a different one?",
					flags.GetProject(c.Scope),
					environment.GetEnvironmentId(),
				),
				false)
			if err != nil {
				return "", err
			}

			if !selectNew {
				return environment.GetEnvironmentId(), nil
			}
		} else {
			return environment.GetEnvironmentId(), nil
		}
	}
}

func (c *Cmd) promptChanges(
	ctx context.Context, to *platformv1.CapsuleSpec,
	changes *field.Diff,
) (*field.Diff, error) {
	text, err := getSpecDiffView(changes)
	if err != nil {
		return nil, err
	}

	list, err := getChangesList(changes)
	if err != nil {
		return nil, err
	}

	errChan := make(chan error)
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()

	grid := tview.NewGrid().SetRows(-1).SetColumns(-1, -2)
	grid.AddItem(text, 0, 1, 1, 1, 0, 0, false)
	grid.AddItem(list, 0, 0, 1, 1, 0, 0, true)

	app := tview.NewApplication().SetRoot(grid, true)
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyBackspace2, tcell.KeyBackspace:
			// remove the last item
			if list.GetCurrentItem() == list.GetItemCount()-1 {
				changes.Changes = changes.Changes[:list.GetCurrentItem()]
			} else {
				changes.Changes = append(changes.Changes[:list.GetCurrentItem()], changes.Changes[list.GetCurrentItem()+1:]...)
			}

			list.RemoveItem(list.GetCurrentItem())

			if len(changes.Changes) == 0 {
				errChan <- errors.CanceledErrorf("No changes left to apply")
			}

			from, err := field.Apply(to, changes.Changes)
			if err != nil {
				errChan <- err
			}

			changes, err = field.Compare(to, from)
			if err != nil {
				errChan <- err
			}

			grid.RemoveItem(text)
			text, err = getSpecDiffView(changes)
			if err != nil {
				errChan <- err
			}
			grid.AddItem(text, 0, 1, 1, 1, 0, 0, false)

			return nil
		case tcell.KeyEsc:
			cancel()
		case tcell.KeyCtrlC:
			errChan <- errors.CanceledErrorf("canceled by user")
		}
		return event
	})

	go func() {
		err := app.Run()
		if err != nil {
			errChan <- err
		}
	}()

	defer app.Stop()

	for {
		select {
		case err := <-errChan:
			return nil, err
		case <-ctx.Done():
			return changes, nil
		}
	}
}

func getSpecDiffView(changes *field.Diff) (*tview.TextView, error) {
	var out bytes.Buffer
	hr := &dyff.HumanReport{
		Report:     *changes.Report,
		OmitHeader: true,
	}
	if err := hr.WriteReport(tview.ANSIWriter(&out)); err != nil {
		return nil, err
	}

	text := tview.NewTextView()
	text.SetTitle(fmt.Sprintf("%s (ESC to continue and CTRL+C to cancel)", "Capsule Diff"))
	text.SetBorder(true)
	text.SetDynamicColors(true)
	text.SetWrap(true)
	text.SetText(out.String())
	text.SetBackgroundColor(tcell.ColorNone)

	return text, nil
}

func getChangesList(changes *field.Diff) (*tview.List, error) {
	list := tview.NewList().ShowSecondaryText(false)
	for _, change := range changes.Changes {
		list.AddItem(change.String(), "", 0, nil)
	}

	list.SetBorder(true).
		SetTitle("Changes (BS to remove selected change)")

	return list, nil
}

func (c *Cmd) getLatestRollout(
	ctx context.Context,
	capsuleID, environmentID, projectID string,
) (*capsule.Rollout, error) {
	r, err := c.Rig.Capsule().ListRollouts(ctx, connect.NewRequest(&capsule.ListRolloutsRequest{
		CapsuleId: capsuleID,
		Pagination: &model.Pagination{
			Offset:     0,
			Limit:      1,
			Descending: true,
		},
		ProjectId:     projectID,
		EnvironmentId: environmentID,
	}))
	if err != nil {
		return nil, err
	}

	for _, rollout := range r.Msg.GetRollouts() {
		return rollout, nil
	}

	return nil, errors.NotFoundErrorf("no rollout for capsule")
}
