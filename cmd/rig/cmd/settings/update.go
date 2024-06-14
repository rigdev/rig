package settings

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/environment"
	settings_api "github.com/rigdev/rig-go-api/api/v1/settings"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
)

const abortedErrMsg = "prompt aborted"

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

	var updates []*settings_api.Update
	for {
		i, _, err := c.Prompter.Select("Select the setting to update (CTRL + C to cancel)",
			[]string{"Notification Notifer", "Git store", "Done"})
		if err != nil {
			if err.Error() == abortedErrMsg {
				return nil
			}
			return err
		}

		done := false
		switch i {
		case 0:
			us, err := c.updateNotificationNotifiers(s)
			if err != nil {
				if err.Error() == abortedErrMsg {
					continue
				}
				return err
			}
			updates = append(updates, us...)
		case 1:
			us, err := c.updateGitStore(s)
			if err != nil {
				if err.Error() == abortedErrMsg {
					continue
				}
				return err
			}

			updates = append(updates, us...)
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

func (c *Cmd) updateNotificationNotifiers(s *settings_api.Settings) ([]*settings_api.Update, error) {
	if len(s.NotificationNotifiers) == 0 {
		fmt.Println("No notification notifiers configured - Let's configure one!")
		notifier := &settings_api.NotificationNotifier{}
		if err := c.UpdateNotifier(notifier); err != nil {
			return nil, err
		}

		s.NotificationNotifiers = append(s.NotificationNotifiers, notifier)
		return []*settings_api.Update{
			{
				Field: &settings_api.Update_SetNotificationNotifiers_{
					SetNotificationNotifiers: &settings_api.Update_SetNotificationNotifiers{
						Notifiers: s.GetNotificationNotifiers(),
					},
				},
			},
		}, nil
	}

	header := []string{"Target Type", "Target ID", "Target Details", "Topics", "Environments"}
	var notifierRows [][]string
	for _, n := range s.GetNotificationNotifiers() {
		notifierRows = append(notifierRows, notifierToRow(n))
	}

	notifierRows = append(notifierRows, []string{"Add new notifier", "", "", "", ""})
	notifierRows = append(notifierRows, []string{"Delete a notifier", "", "", "", ""})
	notifierRows = append(notifierRows, []string{"Done", "", "", "", ""})

	for {
		i, err := c.Prompter.TableSelect("Select the notifier to update (CTRL + C to cancel)", notifierRows, header)
		if err != nil {
			return nil, err
		}

		if i == len(notifierRows)-1 {
			break
		}

		if i == len(notifierRows)-2 {
			// Delete notifier
			i, err := c.Prompter.TableSelect("Select the notifier to delete (CTRL + C to cancel)",
				notifierRows[:len(notifierRows)-3], header)
			if err != nil {
				if err.Error() == abortedErrMsg {
					continue
				}
				return nil, err
			}

			// Remove the notifier from the row list and update the settings
			notifierRows = append(notifierRows[:i], notifierRows[i+1:]...)
			s.NotificationNotifiers = append(s.NotificationNotifiers[:i], s.NotificationNotifiers[i+1:]...)
			continue
		}

		if i == len(notifierRows)-3 {
			// Add new notifier
			notifier := &settings_api.NotificationNotifier{}
			if err := c.UpdateNotifier(notifier); err != nil {
				if err.Error() == abortedErrMsg {
					continue
				}

				return nil, err
			}

			// Add the notifier to the row list and update the settings
			notifierRows = append(notifierRows[:len(notifierRows)-3],
				append([][]string{notifierToRow(notifier)}, notifierRows[len(notifierRows)-3:]...)...)

			s.NotificationNotifiers = append(s.NotificationNotifiers, notifier)
			continue
		}

		// Update existing notifier
		if err := c.UpdateNotifier(s.NotificationNotifiers[i]); err != nil {
			if err.Error() == abortedErrMsg {
				continue
			}

			return nil, err
		}
		notifierRows[i] = notifierToRow(s.NotificationNotifiers[i])
	}

	return []*settings_api.Update{
		{
			Field: &settings_api.Update_SetNotificationNotifiers_{
				SetNotificationNotifiers: &settings_api.Update_SetNotificationNotifiers{
					Notifiers: s.GetNotificationNotifiers(),
				},
			},
		},
	}, nil
}

func notifierToRow(n *settings_api.NotificationNotifier) []string {
	row := []string{}
	if email := n.GetTarget().GetEmail(); email != nil {
		detail := fmt.Sprintf("From: %s", email.GetFromEmail())
		row = append(row, []string{"Email", email.GetId(), detail}...)
	} else if slack := n.GetTarget().GetSlack(); slack != nil {
		detail := fmt.Sprintf("Channel: %s", slack.GetChannelId())
		row = append(row, []string{"Slack", slack.GetWorkspace(), detail}...)
	}

	var topics []string
	for _, t := range n.GetTopics() {
		topics = append(topics, topicToString(t))
	}

	row = append(row, fmt.Sprintf("%v", strings.Join(topics, ", ")), environmentToString(n.GetEnvironments()))
	return row
}

func environmentToString(filter *model.EnvironmentFilter) string {
	if filter == nil {
		return "All"
	}

	if filter.GetAll() != nil {
		str := "All"
		if filter.GetAll().GetIncludeEphemeral() {
			str += " (+ Ephemeral)"
		}

		return str
	}

	if envs := filter.GetSelected(); envs != nil {
		return strings.Join(envs.GetEnvironmentIds(), ", ")
	}

	return "Unknown"
}

func topicToString(t settings_api.NotificationTopic) string {
	switch t {
	case settings_api.NotificationTopic_NOTIFICATION_TOPIC_ISSUE:
		return "Issue"
	case settings_api.NotificationTopic_NOTIFICATION_TOPIC_ROLLOUT:
		return "Rollout"
	default:
		return "Unknown"
	}
}

func (c *Cmd) UpdateNotifier(n *settings_api.NotificationNotifier) error {
	fields := []string{
		"Target",
		"Topics",
		"Environments",
		"Done",
	}

	for {
		i, _, err := c.Prompter.Select("Select the field to update (CTRL + c to cancel)", fields)
		if err != nil {
			return err
		}

		switch i {
		case 0:
			if n.GetTarget() == nil {
				n.Target = &settings_api.NotificationTarget{}
			}

			if err := c.updateNotifierTarget(n.GetTarget()); err != nil {
				if err.Error() == abortedErrMsg {
					continue
				}
				return err
			}
		case 1:
			if err := c.updateNotifierTopics(n); err != nil {
				if err.Error() == abortedErrMsg {
					continue
				}
				return err
			}
		case 2:
			if n.GetEnvironments() == nil {
				n.Environments = &model.EnvironmentFilter{}
			}

			if err := c.updateEnvironmentFilter(n.GetEnvironments()); err != nil {
				if err.Error() == abortedErrMsg {
					continue
				}
				return err
			}
		default:
			return nil
		}
	}
}

func (c *Cmd) updateEnvironmentFilter(filter *model.EnvironmentFilter) error {
	if filter == nil {
		filter = &model.EnvironmentFilter{}
	}

	envResp, err := c.Rig.Environment().List(context.Background(), connect.NewRequest(&environment.ListRequest{}))
	if err != nil {
		return err
	}

	for {
		var envs []string
		for _, e := range envResp.Msg.GetEnvironments() {
			env := e.GetEnvironmentId()
			if slices.Contains(filter.GetSelected().GetEnvironmentIds(), e.GetEnvironmentId()) {
				env += " *"
			}

			envs = append(envs, env)
		}

		all := "All"
		allEphemeral := "All + Ephemeral"
		if filter.GetAll() != nil {
			if filter.GetAll().GetIncludeEphemeral() {
				allEphemeral += " *"
			} else {
				all += " *"
			}
		}

		envs = append(envs, all, allEphemeral, "Done")

		i, _, err := c.Prompter.Select("Select Environments (select current environments marked by * to remove)", envs)
		if err != nil {
			return err
		}

		if i == len(envs)-1 {
			break
		}

		if i == len(envs)-2 {
			if filter.GetAll() == nil {
				filter.Filter = &model.EnvironmentFilter_All_{
					All: &model.EnvironmentFilter_All{},
				}
			}

			filter.GetAll().IncludeEphemeral = true
		} else if i == len(envs)-3 {
			if filter.GetAll() == nil {
				filter.Filter = &model.EnvironmentFilter_All_{
					All: &model.EnvironmentFilter_All{},
				}
			}

			filter.GetAll().IncludeEphemeral = false
		} else {
			env := envResp.Msg.GetEnvironments()[i]

			if filter.GetSelected() == nil {
				filter.Filter = &model.EnvironmentFilter_Selected_{
					Selected: &model.EnvironmentFilter_Selected{},
				}
			}

			if i := slices.Index(filter.GetSelected().GetEnvironmentIds(), env.GetEnvironmentId()); i != -1 {
				filter.GetSelected().EnvironmentIds = slices.Delete(filter.GetSelected().GetEnvironmentIds(), i, i+1)
			} else {
				filter.GetSelected().EnvironmentIds = append(filter.GetSelected().GetEnvironmentIds(), env.GetEnvironmentId())
			}
		}

	}

	return nil
}

func (c *Cmd) updateNotifierTopics(n *settings_api.NotificationNotifier) error {
	availableTopics := []settings_api.NotificationTopic{
		settings_api.NotificationTopic_NOTIFICATION_TOPIC_ISSUE,
		settings_api.NotificationTopic_NOTIFICATION_TOPIC_ROLLOUT,
	}

	for {
		var ts []string
		for _, t := range availableTopics {
			tString := topicToString(t)
			if slices.Contains(n.Topics, t) {
				tString += " *"
			}

			ts = append(ts, tString)
		}
		ts = append(ts, "Done")

		i, _, err := c.Prompter.Select("Select topics (select current topics marked by * to remove)", ts)
		if err != nil {
			return err
		}

		if i == len(ts)-1 {
			break
		}

		t := availableTopics[i]
		if i := slices.Index(n.GetTopics(), t); i != -1 {
			n.Topics = slices.Delete(n.GetTopics(), i, i+1)
		} else {
			n.Topics = append(n.GetTopics(), t)
		}
	}

	return nil
}

func (c *Cmd) updateNotifierTarget(n *settings_api.NotificationTarget) error {
	currentTarget := ""
	switch n.GetTarget().(type) {
	case *settings_api.NotificationTarget_Email:
		currentTarget = "Email"
	case *settings_api.NotificationTarget_Slack:
		currentTarget = "Slack"
	}

	label := "Select the target Type"
	if currentTarget != "" {
		label += fmt.Sprintf(" (Current: %s)", currentTarget)
	}
	i, _, err := c.Prompter.Select(label, []string{"Email", "Slack"})
	if err != nil {
		return err
	}

	switch i {
	case 0:
		target := n.GetEmail()
		if target == nil {
			target = &settings_api.NotificationTarget_EmailTarget{}
		}

		if err := c.updateEmailNotifier(target); err != nil {
			return err
		}

		n.Target = &settings_api.NotificationTarget_Email{
			Email: target,
		}
	case 1:
		slack := n.GetSlack()
		if slack == nil {
			slack = &settings_api.NotificationTarget_SlackTarget{}
		}

		if err := c.updateSlackNotifier(n.GetSlack()); err != nil {
			return err
		}

		n.Target = &settings_api.NotificationTarget_Slack{
			Slack: slack,
		}
	}

	return nil
}

func (c *Cmd) updateSlackNotifier(n *settings_api.NotificationTarget_SlackTarget) error {
	label := "Select the Slack workspace ID"
	if n.GetWorkspace() != "" {
		label += fmt.Sprintf(" (Current: %s)", n.GetWorkspace())
	}

	conf, err := c.Rig.Settings().GetConfiguration(
		context.Background(), connect.NewRequest(&settings_api.GetConfigurationRequest{}),
	)
	if err != nil {
		return err
	}

	// Get the slack workspaces
	var workspaces []string
	for _, w := range conf.Msg.GetConfiguration().GetClient().GetSlack().GetWorkspace() {
		workspaces = append(workspaces, w.GetName())
	}

	if len(workspaces) == 0 {
		return errors.New("no Slack workspaces configured")
	}

	_, w, err := c.Prompter.Select(label, workspaces)
	if err != nil {
		return err
	}

	channelID, err := c.Prompter.Input("Enter a Slack channel ID",
		common.ValidateNonEmptyOpt, common.InputDefaultOpt(n.GetChannelId()))
	if err != nil {
		return err
	}

	if n == nil {
		n = &settings_api.NotificationTarget_SlackTarget{}
	}

	n.Workspace = w
	n.ChannelId = channelID
	return nil
}

func (c *Cmd) updateEmailNotifier(e *settings_api.NotificationTarget_EmailTarget) error {
	label := "Select a configured Email provider"
	if e.GetId() != "" {
		label += fmt.Sprintf(" (Current: %s)", e.GetId())
	}

	conf, err := c.Rig.Settings().GetConfiguration(
		context.Background(), connect.NewRequest(&settings_api.GetConfigurationRequest{}),
	)
	if err != nil {
		return err
	}

	// Get the email providers
	var providers []string
	for _, p := range conf.Msg.GetConfiguration().GetClient().GetEmail() {
		if p.GetType() == settings_api.EmailType_EMAIL_TYPE_MAILJET {
			providers = append(providers, fmt.Sprintf("%s (Mailjet)", p.GetId()))
		} else if p.GetType() == settings_api.EmailType_EMAIL_TYPE_SMTP {
			providers = append(providers, fmt.Sprintf("%s (SMTP)", p.GetId()))
		}
	}

	if len(providers) == 0 {
		return errors.New("no email providers configured")
	}

	_, id, err := c.Prompter.Select(label, providers)
	if err != nil {
		return err
	}

	fromEmail, err := c.Prompter.Input("Enter the email address to send from",
		common.ValidateEmailOpt, common.InputDefaultOpt(e.GetFromEmail()))
	if err != nil {
		return err
	}

	if e == nil {
		e = &settings_api.NotificationTarget_EmailTarget{}
	}

	e.Id = id
	e.FromEmail = fromEmail

	return nil
}

func (c *Cmd) updateGitStore(s *settings_api.Settings) ([]*settings_api.Update, error) {
	gitStore := s.GetGitStore()
	if gitStore == nil {
		gitStore = &model.GitStore{}
	}

	fields := []string{
		"Disabled",
		"Repository",
		"Branch",
		"Capsule Path",
		"Commit Template",
		"Environments",
		"Done",
	}

	for {
		i, _, err := c.Prompter.Select("Select the field to update (CTRL + c to cancel)", fields)
		if err != nil {
			return nil, err
		}

		switch i {
		case 0:
			disable, err := c.Prompter.Confirm("Disable Git store", false)
			if err != nil {
				if err.Error() == abortedErrMsg {
					continue
				}
				return nil, err
			}

			gitStore.Disabled = disable
		case 1:
			repo, err := c.Prompter.Input("Enter the repository URL",
				common.ValidateNonEmptyOpt, common.InputDefaultOpt(gitStore.GetRepository()))
			if err != nil {
				if err.Error() == abortedErrMsg {
					continue
				}
				return nil, err
			}

			gitStore.Repository = repo
		case 2:
			branch, err := c.Prompter.Input("Enter the branch",
				common.ValidateNonEmptyOpt, common.InputDefaultOpt(gitStore.GetBranch()))
			if err != nil {
				if err.Error() == abortedErrMsg {
					continue
				}
				return nil, err
			}

			gitStore.Branch = branch
		case 3:
			path, err := c.Prompter.Input("Enter the capsule path",
				common.ValidateNonEmptyOpt, common.InputDefaultOpt(gitStore.GetCapsulePath()))
			if err != nil {
				if err.Error() == abortedErrMsg {
					continue
				}
				return nil, err
			}

			gitStore.CapsulePath = path
		case 4:
			template, err := c.Prompter.Input("Enter the commit template",
				common.ValidateNonEmptyOpt, common.InputDefaultOpt(gitStore.GetCommitTemplate()))
			if err != nil {
				if err.Error() == abortedErrMsg {
					continue
				}
				return nil, err
			}

			gitStore.CommitTemplate = template
		case 5:
			if gitStore.Environments == nil {
				gitStore.Environments = &model.EnvironmentFilter{}
			}

			if err := c.updateEnvironmentFilter(gitStore.GetEnvironments()); err != nil {
				if err.Error() == abortedErrMsg {
					continue
				}
				return nil, err
			}
		default:
			return []*settings_api.Update{
				{
					Field: &settings_api.Update_SetGitStore{
						SetGitStore: gitStore,
					},
				},
			}, nil
		}
	}
}
