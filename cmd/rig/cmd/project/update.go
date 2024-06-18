package project

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/environment"
	"github.com/rigdev/rig-go-api/api/v1/project"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
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

	envResp, err := c.Rig.Environment().List(ctx, connect.NewRequest(&environment.ListRequest{
		ProjectFilter: p.GetProjectId(),
	}))
	if err != nil {
		return err
	}

	var updates []*project.Update
	for {
		i, _, err := c.Prompter.Select("Select the setting to update (CTRL + C to cancel)",
			[]string{"Git store", "Done"})
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
