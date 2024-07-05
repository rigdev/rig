package project

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/environment"
	"github.com/rigdev/rig-go-api/api/v1/project"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
)

func (c *Cmd) update(ctx context.Context, _ *cobra.Command, _ []string) error {
	if !c.Scope.IsInteractive() {
		return errors.New("non-interactive mode is not supported for this command")
	}
	projectID := flags.Flags.Project
	if projectID == "" {
		resp, err := c.Rig.Project().List(ctx, connect.NewRequest(&project.ListRequest{}))
		if err != nil {
			return err
		}
		var projects []string
		for _, p := range resp.Msg.GetProjects() {
			projects = append(projects, p.GetProjectId())
		}
		_, projectID, err = c.Prompter.Select("Project", projects)
		if err != nil {
			return err
		}
	}

	resp, err := c.Rig.Project().Get(ctx, connect.NewRequest(&project.GetRequest{
		ProjectId: projectID,
	}))
	if err != nil {
		return err
	}
	p := resp.Msg.GetProject()

	resp2, err := c.Rig.Project().GetEffectiveGitSettings(ctx, connect.NewRequest(&project.GetEffectiveGitSettingsRequest{
		ProjectId: projectID,
	}))
	if err != nil {
		return err
	}
	p.GitStore = resp2.Msg.GetGit()

	envResp, err := c.Rig.Environment().List(ctx, connect.NewRequest(&environment.ListRequest{
		ProjectFilter: p.GetProjectId(),
	}))
	if err != nil {
		return err
	}

	var updates []*project.Update
	for {
		i, _, err := c.Prompter.Select("Select the setting to update (CTRL + C to cancel)",
			[]string{"Git store", "Notifiers", "Done"})
		if err != nil {
			if common.ErrIsAborted(err) {
				return nil
			}
			return err
		}

		done := false
		switch i {
		case 0:
			p.GitStore, err = common.PromptGitStore(c.Prompter, p.GetGitStore(), envResp.Msg.GetEnvironments())
			if err != nil {
				if common.ErrIsAborted(err) {
					continue
				}
				return err
			}

			updates = append(updates, &project.Update{
				Field: &project.Update_SetGitStore{
					SetGitStore: p.GetGitStore(),
				},
			})
		case 1:
			if err := c.updateNotifiers(ctx, p.GetNotifiers()); err != nil {
				if common.ErrIsAborted(err) {
					continue
				}
				return err
			}

			updates = append(updates, &project.Update{
				Field: &project.Update_Notifiers{
					Notifiers: p.GetNotifiers(),
				},
			})
		case 2:
			done = true
		}
		if done {
			break
		}
	}

	if len(updates) == 0 {
		fmt.Println("No updates to make")
		return nil
	}

	if _, err = c.Rig.Project().Update(ctx, connect.NewRequest(&project.UpdateRequest{
		Updates:   updates,
		ProjectId: projectID,
	})); err != nil {
		return err
	}

	fmt.Println("Project updated")
	return nil
}

func (c *Cmd) updateNotifiers(ctx context.Context, p *project.NotificationNotifiers) error {
	if p == nil {
		p = &project.NotificationNotifiers{}
	}

	enableDisableStr := "disable global notifiers"
	if p.GetDisabled() {
		enableDisableStr = "enable global notifiers"
	}

	i, _, err := c.Prompter.Select("Select the field to update (CTRL + c to cancel)", []string{
		enableDisableStr,
		"Update Notifiers",
	})
	if err != nil {
		return err
	}

	switch i {
	case 0:
		p.Disabled = !p.GetDisabled()
	case 1:
		notifiers, err := common.PromptNotificationNotifiers(ctx, c.Prompter, c.Rig, p.GetNotifiers())
		if err != nil {
			return err
		}

		p.Notifiers = notifiers
	}

	return nil
}

func (c *Cmd) updateGit(ctx context.Context, _ *cobra.Command, _ []string) error {
	gitStore := &model.GitStore{}
	if resp, err := c.Rig.Project().GetEffectiveGitSettings(ctx, connect.NewRequest(&project.GetEffectiveGitSettingsRequest{
		ProjectId: flags.GetProject(c.Scope),
	})); errors.IsNotFound(err) {
	} else if err != nil {
		return err
	} else {
		gitStore = resp.Msg.GetGit()
	}
	if err := common.UpdateGit(ctx, c.Rig, gitFlags, c.Scope.IsInteractive(), c.Prompter, gitStore); err != nil {
		return err
	}

	if _, err := c.Rig.Project().Update(ctx, connect.NewRequest(&project.UpdateRequest{
		Updates:   []*project.Update{{Field: &project.Update_SetGitStore{SetGitStore: gitStore}}},
		ProjectId: flags.GetProject(c.Scope),
	})); err != nil {
		return err
	}

	fmt.Println("Updated project git store settings to:\n")
	if err := common.FormatPrint(gitStore, common.OutputTypeYAML); err != nil {
		return err
	}

	return nil
}
