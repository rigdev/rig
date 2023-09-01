package auth

import (
	"context"
	"reflect"

	"github.com/rigdev/rig-go-api/api/v1/authentication"
	api_user "github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/hash"
	"github.com/rigdev/rig/pkg/uuid"
)

func (s *Service) LoginUserPassword(ctx context.Context, id *model.UserIdentifier, pw string) (uuid.UUID, *model.UserInfo, *authentication.Token, error) {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return uuid.Nil, nil, nil, err
	}

	u, err := s.us.GetUserByIdentifier(ctx, id)
	if err != nil {
		return uuid.Nil, nil, nil, err
	}

	userID := uuid.UUID(u.GetUserId())

	p, err := s.ps.GetProject(ctx)
	if err != nil {
		return uuid.Nil, nil, nil, err
	}

	set, err := s.us.GetSettings(ctx)
	if err != nil {
		return uuid.Nil, nil, nil, err
	}

	userPW, err := s.us.GetPassword(ctx, userID)
	if err != nil {
		return uuid.Nil, nil, nil, err
	}

	h := hash.New(userPW.GetConfig())
	if err := h.Compare(pw, userPW); err != nil {
		return uuid.Nil, nil, nil, err
	}

	var lt model.LoginType
	switch v := id.GetIdentifier().(type) {
	case *model.UserIdentifier_Email:
		lt = model.LoginType_LOGIN_TYPE_EMAIL_PASSWORD
		if set.GetIsVerifiedEmailRequired() && !u.GetIsEmailVerified() {
			if err := s.sendVerificationEmail(ctx, userID, u, p, set); err != nil {
				return uuid.Nil, nil, nil, err
			}
			return uuid.Nil, nil, nil, errors.FailedPreconditionErrorf("email is not verified")
		}
	case *model.UserIdentifier_PhoneNumber:
		lt = model.LoginType_LOGIN_TYPE_PHONE_PASSWORD
		if set.GetIsVerifiedPhoneRequired() && !u.GetIsPhoneVerified() {
			if err := s.sendVerificationText(ctx, userID, u, p, set); err != nil {
				return uuid.Nil, nil, nil, err
			}
			return uuid.Nil, nil, nil, errors.FailedPreconditionErrorf("phone number is not verified")
		}
	case *model.UserIdentifier_Username:
		lt = model.LoginType_LOGIN_TYPE_USERNAME_PASSWORD
	default:
		return uuid.Nil, nil, nil, errors.InvalidArgumentErrorf("invalid login type '%'", reflect.TypeOf(v))
	}

	sessionID, ss, err := s.newSession(ctx, userID, &api_user.AuthMethod{Method: &api_user.AuthMethod_LoginType{LoginType: lt}})
	if err != nil {
		return uuid.Nil, nil, nil, err
	}

	token, err := s.generateToken(ctx, sessionID, ss, projectID, userID, auth.SubjectTypeUser, nil, set)
	if err != nil {
		return uuid.Nil, nil, nil, err
	}

	return userID, u.GetUserInfo(), token, nil
}

func (s *Service) Logout(ctx context.Context) error {
	c, err := auth.GetClaims(ctx)
	if err != nil {
		return err
	}

	sessionID := c.GetSessionID()
	if sessionID.IsNil() {
		return errors.InvalidArgumentErrorf("no session ID in JWT token")
	}

	if err := s.deleteSession(ctx, c.GetSubject(), sessionID); errors.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	return nil
}
