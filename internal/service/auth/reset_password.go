package auth

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"time"

	"github.com/rigdev/rig-go-api/api/v1/project/settings"
	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/pkg/gateway/email"
	"github.com/rigdev/rig/pkg/crypto"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/hash"
	"github.com/rigdev/rig/pkg/utils"
	"github.com/rigdev/rig/pkg/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Service) ResetPassword(ctx context.Context, identifier *model.UserIdentifier, code, newPassword string) error {
	u, err := s.us.GetUserByIdentifier(ctx, identifier)
	if err != nil {
		return err
	}

	userID := uuid.UUID(u.GetUserId())

	vc, err := s.vcr.Get(ctx, userID, user.VerificationType_VERIFICATION_TYPE_RESET_PASSWORD)
	if errors.IsNotFound(err) {
		return errors.InvalidArgumentErrorf("invalid verification code")
	} else if err != nil {
		return err
	} else if vc.GetExpiresAt().AsTime().Before(time.Now()) {
		return errors.InvalidArgumentErrorf("verification code expired")
	}

	h := hash.New(vc.GetCode().GetConfig())
	if err := h.Compare(code, vc.GetCode()); err != nil {
		return errors.InvalidArgumentErrorf("invalid verification code")
	}

	if err := s.us.UpdateUser(ctx, userID, []*user.Update{{
		Field: &user.Update_Password{Password: newPassword},
	}}); err != nil {
		return err
	}

	return s.vcr.Delete(ctx, userID, user.VerificationType_VERIFICATION_TYPE_RESET_PASSWORD)
}

func (s *Service) SendPasswordReset(ctx context.Context, identifier *model.UserIdentifier) error {
	u, err := s.us.GetUserByIdentifier(ctx, identifier)
	if err != nil {
		return err
	}

	us, err := s.us.GetSettings(ctx)
	if err != nil {
		return err
	}

	code, err := crypto.GenerateSymmetricKey(6, crypto.Numeric)
	if err != nil {
		return err
	}

	pw, err := hash.New(us.GetPasswordHashing()).Generate(code)
	if err != nil {
		return err
	}

	vc := &user.VerificationCode{
		Code:      pw,
		UserId:    u.GetUserId(),
		Type:      user.VerificationType_VERIFICATION_TYPE_RESET_PASSWORD,
		ExpiresAt: timestamppb.New(time.Now().Add(us.GetVerificationCodeTtl().AsDuration())),
	}

	if _, err := s.vcr.Create(ctx, vc); err != nil {
		return err
	}

	switch identifier.GetIdentifier().(type) {
	case *model.UserIdentifier_Username:
		return errors.FailedPreconditionErrorf("cannot send reset password email, no email address available for user")
	case *model.UserIdentifier_Email:
		ep, err := s.ps.GetEmailProvider(ctx)
		if errors.IsNotFound(err) {
			s.logger.Warn("cannot send reset password email, no email provider available for project")
			return nil
		} else if err != nil {
			return err
		}

		projectSettings, err := s.ps.GetProjectSettings(ctx)
		if err != nil {
			return err
		}

		msg, err := GetResetPasswordEmail(projectSettings, u, code)
		if err != nil {
			return err
		}
		s.logger.Info("sending reset password email", zap.String("email", u.GetUserInfo().GetEmail()))
		return ep.Send(ctx, msg)
	case *model.UserIdentifier_PhoneNumber:
		return errors.FailedPreconditionErrorf("cannot send reset password email, no email address available for user")
	}
	return nil
}

func GetResetPasswordEmail(set *settings.Settings, u *user.User, code string) (*email.Message, error) {
	sub := _resetPasswordEmailSubject
	if s := set.GetTemplates().GetResetPasswordEmail().GetSubject(); s != "" {
		sub = s
	}

	body := _resetPasswordBody
	if s := set.GetTemplates().GetResetPasswordEmail().GetBody(); s != "" {
		body = s
	}

	t, err := template.New("reset_password_email").Parse(body)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	msg := &email.Message{
		To: []*email.Recipient{{
			Email: u.GetUserInfo().GetEmail(),
			Name:  fmt.Sprint(u.GetProfile().GetFirstName(), " ", u.GetProfile().GetLastName()),
		}},
		Subject:  sub,
		HTMLPart: buffer.String(),
	}
	return msg, nil
}

const (
	_resetPasswordEmailSubject = "Reset your password"
	_resetPasswordBody         = `<h2>Reset your password</h2><p> Enter the following code to reset your password: {{ .Code }}</p>`
)
