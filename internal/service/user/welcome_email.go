package user

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	project_settings "github.com/rigdev/rig-go-api/api/v1/project/settings"
	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig/internal/gateway/email"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/utils"
	"github.com/rigdev/rig/pkg/uuid"
	"go.uber.org/zap"
)

func (s *service) sendWelcomeEmail(ctx context.Context, userID uuid.UUID, u *user.User) error {
	if u.GetUserInfo().GetEmail() == "" {
		return nil
	}

	ep, err := s.ps.GetEmailProvider(ctx)
	if errors.IsNotFound(err) {
		s.logger.Warn("cannot send verification email, no email provider available for project")
		return nil
	} else if err != nil {
		return err
	}

	projectSetting, err := s.ps.GetProjectSettings(ctx)
	if err != nil {
		return err
	}

	sub, body := getWelcomeEmail(projectSetting)
	t, err := template.New("verification_email").Parse(body)
	if err != nil {
		return err
	}

	var buffer bytes.Buffer
	if err := t.Execute(&buffer, struct {
		Email      string
		Identifier string
	}{
		Email:      u.GetUserInfo().GetEmail(),
		Identifier: utils.UserName(u),
	}); err != nil {
		return err
	}

	msg := &email.Message{
		To: []*email.Recipient{{
			Email: u.GetUserInfo().GetEmail(),
			Name:  fmt.Sprint(u.GetProfile().GetFirstName(), " ", u.GetProfile().GetLastName()),
		}},
		Subject:  sub,
		HTMLPart: buffer.String(),
	}

	s.logger.Info("sending welcome email", zap.String("email", u.GetUserInfo().GetEmail()))
	if err := ep.Send(ctx, msg); err != nil {
		return err
	}

	return nil
}

func getWelcomeEmail(set *project_settings.Settings) (string, string) {
	subject := _welcomeEmailSubject
	if s := set.GetTemplates().GetWelcomeEmail().GetSubject(); s != "" {
		subject = s
	}

	body := _welcomeEmailBody
	if s := set.GetTemplates().GetWelcomeEmail().GetBody(); s != "" {
		body = s
	}

	return subject, body
}

const (
	_welcomeEmailSubject = "Welcome to Rig!"
	_welcomeEmailBody    = `<h2>Welcome to Rig!</h2><p>Login to get started</p>`
)
