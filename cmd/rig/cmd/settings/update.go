package settings

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/environment"
	settings_api "github.com/rigdev/rig-go-api/api/v1/settings"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
)

func (c *Cmd) update(ctx context.Context, cmd *cobra.Command, _ []string) error {
	if !c.Scope.IsInteractive() {
		return errors.New("non-interactive mode is not supported for this command")
	}

	settingsResp, err := c.Rig.Settings().GetSettings(ctx, connect.NewRequest(&settings_api.GetSettingsRequest{}))
	if err != nil {
		return err
	}

	s := settingsResp.Msg.GetSettings()
	if s == nil {
		s = &settings_api.Settings{}
	}

	envResp, err := c.Rig.Environment().List(ctx, connect.NewRequest(&environment.ListRequest{}))
	if err != nil {
		return err
	}

	var updates []*settings_api.Update
	for {
		i, _, err := c.Prompter.Select("Select the setting to update (CTRL + C to cancel)",
			[]string{"Notification Notifer", "Git store", "Done"})
		if err != nil {
			if common.ErrIsAborted(err) {
				return nil
			}
			return err
		}

		done := false
		switch i {
		case 0:
			notifiers, err := common.PromptNotificationNotifiers(ctx, c.Prompter, c.Rig, s.GetNotificationNotifiers())
			if err != nil {
				if common.ErrIsAborted(err) {
					continue
				}
				return err
			}

			s.NotificationNotifiers = notifiers

			updates = append(updates, &settings_api.Update{
				Field: &settings_api.Update_SetNotificationNotifiers_{
					SetNotificationNotifiers: &settings_api.Update_SetNotificationNotifiers{
						Notifiers: notifiers,
					},
				},
			})
		case 1:
			gitStore, err := common.PromptGitStore(c.Prompter, s.GetGitStore(), envResp.Msg.GetEnvironments())
			if err != nil {
				if common.ErrIsAborted(err) {
					continue
				}
				return err
			}

			updates = append(updates, &settings_api.Update{
				Field: &settings_api.Update_SetGitStore{
					SetGitStore: gitStore,
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
		cmd.Println("No updates to make")
		return nil
	}

	if _, err = c.Rig.Settings().UpdateSettings(ctx, connect.NewRequest(&settings_api.UpdateSettingsRequest{
		Updates: updates,
	})); err != nil {
		return err
	}

	cmd.Println("Settings updated")
	return nil
}

func (c *Cmd) updateGit(ctx context.Context, _ *cobra.Command, _ []string) error {
	resp, err := c.Rig.Settings().GetSettings(ctx, connect.NewRequest(&settings_api.GetSettingsRequest{}))
	if err != nil {
		return err
	}

	gitStore := resp.Msg.GetSettings().GetGitStore()
	if gitStore, err = common.UpdateGit(ctx, c.Rig, gitFlags, c.Scope.IsInteractive(), c.Prompter, gitStore); err != nil {
		return err
	}

	if _, err := c.Rig.Settings().UpdateSettings(ctx, connect.NewRequest(&settings_api.UpdateSettingsRequest{
		Updates: []*settings_api.Update{{
			Field: &settings_api.Update_SetGitStore{
				SetGitStore: gitStore,
			},
		}},
	})); err != nil {
		return err
	}

	fmt.Println("Updated global git store settings to:")
	fmt.Println()
	if err := common.FormatPrint(gitStore, common.OutputTypeYAML); err != nil {
		return err
	}

	return nil
}
