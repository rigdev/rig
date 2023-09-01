package auth

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/project"
	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig-go-api/api/v1/user/settings"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
)

func (s *Service) VerifyPhoneNumber(ctx context.Context, email string, code string) error {
	return errors.UnimplementedErrorf("VerifyPhoneNumber")
}

func (s *Service) sendVerificationText(ctx context.Context, userID uuid.UUID, u *user.User, p *project.Project, set *settings.Settings) error {
	return errors.UnimplementedErrorf("sendVerificationText")
}
