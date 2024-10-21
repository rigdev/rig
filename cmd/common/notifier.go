package common

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/environment"
	"github.com/rigdev/rig-go-api/api/v1/settings"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/pkg/errors"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/proto"
)

func PromptNotificationNotifiers(
	ctx context.Context,
	prompter Prompter,
	rig rig.Client,
	notifiers []*model.NotificationNotifier,
) ([]*model.NotificationNotifier, error) {
	if len(notifiers) == 0 {
		fmt.Println("No notification notifiers configured - Let's configure one!")
		n, err := updateNotifier(ctx, rig, prompter, nil)
		if err != nil {
			return nil, err
		}

		return []*model.NotificationNotifier{
			n,
		}, nil
	}

	header := []string{"Target Type", "Target ID", "Target Details", "Topics", "Environments"}
	var notifierRows [][]string
	for _, n := range notifiers {
		notifierRows = append(notifierRows, notifierToRow(n))
	}

	notifierRows = append(notifierRows, []string{"Add new notifier", "", "", "", ""})
	notifierRows = append(notifierRows, []string{"Delete a notifier", "", "", "", ""})
	notifierRows = append(notifierRows, []string{"Done", "", "", "", ""})

	for {
		i, err := prompter.TableSelect("Select the notifier to update (CTRL + C to cancel)", notifierRows, header)
		if err != nil {
			return nil, err
		}

		if i == len(notifierRows)-1 {
			break
		}

		if i == len(notifierRows)-2 {
			// Delete notifier
			i, err := prompter.TableSelect("Select the notifier to delete (CTRL + C to cancel)",
				notifierRows[:len(notifierRows)-3], header)
			if err != nil {
				if ErrIsAborted(err) {
					continue
				}
				return nil, err
			}

			// Remove the notifier from the row list and update the settings
			notifierRows = append(notifierRows[:i], notifierRows[i+1:]...)
			notifiers = append(notifiers[:i], notifiers[i+1:]...)
			continue
		}

		if i == len(notifierRows)-3 {
			// Add new notifier
			n, err := updateNotifier(ctx, rig, prompter, nil)
			if err != nil {
				if ErrIsAborted(err) {
					continue
				}

				return nil, err
			}

			notifierRows = append(notifierRows[:len(notifierRows)-3],
				append([][]string{notifierToRow(n)}, notifierRows[len(notifierRows)-3:]...)...)

			notifiers = append(notifiers, n)
			continue
		}

		// Update existing notifier
		n, err := updateNotifier(ctx, rig, prompter, notifiers[i])
		if err != nil {
			if ErrIsAborted(err) {
				continue
			}

			return nil, err
		}
		notifiers[i] = n
		notifierRows[i] = notifierToRow(notifiers[i])
	}

	return notifiers, nil
}

func notifierToRow(n *model.NotificationNotifier) []string {
	row := []string{}
	if email := n.GetTarget().GetEmail(); email != nil {
		detail := fmt.Sprintf("%s -> %s", email.GetFromEmail(), strings.Join(email.GetToEmails(), ","))
		row = append(row, []string{"Email", email.GetId(), detail}...)
	} else if slack := n.GetTarget().GetSlack(); slack != nil {
		detail := fmt.Sprintf("Channel: %s", slack.GetChannelId())
		row = append(row, []string{"Slack", slack.GetWorkspace(), detail}...)
	}

	var topics []string
	for _, t := range n.GetTopics() {
		topics = append(topics, TopicToString(t))
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

func TopicToString(t model.NotificationTopic) string {
	switch t {
	case model.NotificationTopic_NOTIFICATION_TOPIC_ISSUE:
		return "Issue"
	case model.NotificationTopic_NOTIFICATION_TOPIC_ROLLOUT:
		return "Rollout"
	case model.NotificationTopic_NOTIFICATION_TOPIC_CAPSULE:
		return "Capsule"
	case model.NotificationTopic_NOTIFICATION_TOPIC_USER:
		return "User"
	case model.NotificationTopic_NOTIFICATION_TOPIC_PROJECT:
		return "Project"
	case model.NotificationTopic_NOTIFICATION_TOPIC_ENVIRONMENT:
		return "Environment"
	default:
		return "Unknown"
	}
}

func updateNotifier(
	ctx context.Context,
	rig rig.Client,
	prompter Prompter,
	n *model.NotificationNotifier,
) (*model.NotificationNotifier, error) {
	if n == nil {
		n = &model.NotificationNotifier{}
	}

	n = proto.Clone(n).(*model.NotificationNotifier)

	fields := []string{
		"Target",
		"Topics",
		"Environments",
		"Done",
	}

	for {
		i, _, err := prompter.Select("Select the field to update (CTRL + c to cancel)", fields)
		if err != nil {
			return nil, err
		}

		switch i {
		case 0:
			if err := updateNotifierTarget(ctx, rig, prompter, n); err != nil {
				if ErrIsAborted(err) {
					continue
				}
				return nil, err
			}
		case 1:
			if err := updateNotifierTopics(prompter, n); err != nil {
				if ErrIsAborted(err) {
					continue
				}
				return nil, err
			}
		case 2:
			if n.GetEnvironments() == nil {
				n.Environments = &model.EnvironmentFilter{}
			}

			resp, err := rig.Environment().List(ctx, connect.NewRequest(&environment.ListRequest{}))
			if err != nil {
				return nil, err
			}
			n.Environments, err = PromptEnvironmentFilter(prompter, n.GetEnvironments(), resp.Msg.GetEnvironments())
			if err != nil {
				if ErrIsAborted(err) {
					continue
				}
				return nil, err
			}
		default:
			return n, nil
		}
	}
}

func updateNotifierTopics(prompter Prompter, n *model.NotificationNotifier) error {
	availableTopics := []model.NotificationTopic{
		model.NotificationTopic_NOTIFICATION_TOPIC_ISSUE,
		model.NotificationTopic_NOTIFICATION_TOPIC_ROLLOUT,
	}

	for {
		var ts []string
		for _, t := range availableTopics {
			tString := TopicToString(t)
			if slices.Contains(n.Topics, t) {
				tString += " *"
			}

			ts = append(ts, tString)
		}
		ts = append(ts, "Done")

		i, _, err := prompter.Select("Select topics (select current topics marked by * to remove)", ts)
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

func updateNotifierTarget(
	ctx context.Context,
	rig rig.Client,
	prompter Prompter,
	notifier *model.NotificationNotifier,
) error {
	if notifier.GetTarget() == nil {
		notifier.Target = &model.NotificationTarget{}
	}

	n := notifier.Target
	currentTarget := ""
	switch n.GetTarget().(type) {
	case *model.NotificationTarget_Email:
		currentTarget = "Email"
	case *model.NotificationTarget_Slack:
		currentTarget = "Slack"
	}

	label := "Select the target Type"
	if currentTarget != "" {
		label += fmt.Sprintf(" (Current: %s)", currentTarget)
	}
	i, _, err := prompter.Select(label, []string{"Email", "Slack"})
	if err != nil {
		return err
	}

	switch i {
	case 0:
		target := n.GetEmail()
		if target == nil {
			target = &model.NotificationTarget_EmailTarget{}
		}

		if err := updateEmailNotifier(ctx, rig, prompter, target); err != nil {
			return err
		}

		n.Target = &model.NotificationTarget_Email{
			Email: target,
		}
	case 1:
		slack := n.GetSlack()
		if slack == nil {
			slack = &model.NotificationTarget_SlackTarget{}
		}

		if err := updateSlackNotifier(ctx, rig, prompter, slack); err != nil {
			return err
		}

		n.Target = &model.NotificationTarget_Slack{
			Slack: slack,
		}
	}

	return nil
}

func updateSlackNotifier(
	ctx context.Context,
	rig rig.Client,
	prompter Prompter,
	n *model.NotificationTarget_SlackTarget,
) error {
	label := "Select the Slack workspace ID"
	if n.GetWorkspace() != "" {
		label += fmt.Sprintf(" (Current: %s)", n.GetWorkspace())
	}

	conf, err := rig.Settings().GetConfiguration(
		ctx, connect.NewRequest(&settings.GetConfigurationRequest{}),
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

	_, w, err := prompter.Select(label, workspaces)
	if err != nil {
		return err
	}

	channelID, err := prompter.Input("Enter a Slack channel ID",
		ValidateNonEmptyOpt, InputDefaultOpt(n.GetChannelId()))
	if err != nil {
		return err
	}

	if n == nil {
		n = &model.NotificationTarget_SlackTarget{}
	}

	n.Workspace = w
	n.ChannelId = channelID
	return nil
}

func updateEmailNotifier(
	ctx context.Context,
	rig rig.Client,
	prompter Prompter,
	e *model.NotificationTarget_EmailTarget,
) error {
	label := "Select a configured Email provider"
	if e.GetId() != "" {
		label += fmt.Sprintf(" (Current: %s)", e.GetId())
	}

	conf, err := rig.Settings().GetConfiguration(
		ctx, connect.NewRequest(&settings.GetConfigurationRequest{}),
	)
	if err != nil {
		return err
	}

	// Get the email providers
	var providers []string
	for _, p := range conf.Msg.GetConfiguration().GetClient().GetEmail() {
		if p.GetType() == settings.EmailType_EMAIL_TYPE_MAILJET {
			providers = append(providers, fmt.Sprintf("%s (Mailjet)", p.GetId()))
		} else if p.GetType() == settings.EmailType_EMAIL_TYPE_SMTP {
			providers = append(providers, fmt.Sprintf("%s (SMTP)", p.GetId()))
		}
	}

	if len(providers) == 0 {
		return errors.New("no email providers configured")
	}

	_, id, err := prompter.Select(label, providers)
	if err != nil {
		return err
	}

	fromEmail, err := prompter.Input("Enter the email address to send from",
		ValidateEmailOpt, InputDefaultOpt(e.GetFromEmail()))
	if err != nil {
		return err
	}

	toEmail, err := prompter.Input("Enter the email address to send to",
		ValidateEmailOpt, InputDefaultOpt(strings.Join(e.GetToEmails(), ",")))
	if err != nil {
		return err
	}

	if e == nil {
		e = &model.NotificationTarget_EmailTarget{}
	}

	e.Id = strings.Split(id, " ")[0]
	e.FromEmail = fromEmail
	e.ToEmails = []string{
		toEmail,
	}

	return nil
}
