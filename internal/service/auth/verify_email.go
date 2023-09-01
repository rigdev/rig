package auth

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"time"

	"github.com/rigdev/rig-go-api/api/v1/project"
	project_settings "github.com/rigdev/rig-go-api/api/v1/project/settings"
	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig-go-api/api/v1/user/settings"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/internal/gateway/email"
	"github.com/rigdev/rig/pkg/crypto"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/hash"
	"github.com/rigdev/rig/pkg/utils"
	"github.com/rigdev/rig/pkg/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Service) VerifyEmail(ctx context.Context, email string, code string) error {
	id := &model.UserIdentifier{
		Identifier: &model.UserIdentifier_Email{Email: email},
	}
	u, err := s.us.GetUserByIdentifier(ctx, id)
	if err != nil {
		return err
	}

	userID := uuid.UUID(u.GetUserId())

	vc, err := s.vcr.Get(ctx, userID, user.VerificationType_VERIFICATION_TYPE_EMAIL)
	if errors.IsNotFound(err) {
		return errors.InvalidArgumentErrorf("invalid verification code")
	} else if err != nil {
		return err
	} else if vc.GetExpiresAt().AsTime().Before(time.Now()) {
		return errors.InvalidArgumentErrorf("verification code expired")
	}

	h := hash.New(vc.GetCode().GetConfig())
	if err := h.Compare(code, vc.GetCode()); err != nil {
		return err
	}

	if err := s.us.UpdateUser(ctx, userID, []*user.Update{{
		Field: &user.Update_IsEmailVerified{IsEmailVerified: true},
	}}); err != nil {
		return err
	}

	return s.vcr.Delete(ctx, userID, user.VerificationType_VERIFICATION_TYPE_EMAIL)
}

func (s *Service) sendVerificationEmail(ctx context.Context, userID uuid.UUID, u *user.User, p *project.Project, set *settings.Settings) error {
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

	vc, err := s.vcr.Get(ctx, userID, user.VerificationType_VERIFICATION_TYPE_EMAIL)
	if errors.IsNotFound(err) {
		// Let's create a new one.
	} else if err != nil {
		return err
	} else if vc.GetExpiresAt().AsTime().After(time.Now()) {
		// We already have an outstanding verification code.
		return nil
	} else {
		s.logger.Debug("ignoring expired verification code")
	}

	code, err := crypto.GenerateSymmetricKey(6, crypto.Numeric)
	if err != nil {
		return err
	}

	pw, err := hash.New(set.GetPasswordHashing()).Generate(code)
	if err != nil {
		return err
	}

	vc = &user.VerificationCode{
		Code:      pw,
		UserId:    userID.String(),
		Type:      user.VerificationType_VERIFICATION_TYPE_EMAIL,
		ExpiresAt: timestamppb.New(time.Now().Add(set.GetVerificationCodeTtl().AsDuration())),
	}
	if _, err := s.vcr.Create(ctx, vc); err != nil {
		return err
	}

	projectSetting, err := s.ps.GetProjectSettings(ctx)
	if err != nil {
		return err
	}

	sub, body := getVerifyEmail(projectSetting)
	t, err := template.New("verification_email").Parse(body)
	if err != nil {
		return err
	}

	var buffer bytes.Buffer
	if err := t.Execute(&buffer, struct {
		Code       string
		Email      string
		Identifier string
	}{
		Code:       code,
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

	s.logger.Info("sending verification email", zap.String("email", u.GetUserInfo().GetEmail()))
	if err := ep.Send(ctx, msg); err != nil {
		return err
	}

	return nil
}

func getVerifyEmail(set *project_settings.Settings) (string, string) {
	subject := _verifyYourEmailSubject
	if s := set.GetTemplates().GetVerifyEmail().GetSubject(); s != "" {
		subject = s
	}

	body := _verifyYourEmailBody
	if s := set.GetTemplates().GetVerifyEmail().GetBody(); s != "" {
		body = s
	}

	return subject, body
}

const (
	_verifyYourEmailSubject = "Confirm your signup"
	_verifyYourEmailBody    = `<h2>Confirm your signup</h2><p> Enter the following code to verify your email: {{ .Code }}</p>`
)
