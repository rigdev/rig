package user

import (
	"context"
	"io"
	"net/mail"
	"reflect"

	"github.com/nyaruka/phonenumbers"
	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig-go-api/api/v1/user/settings"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/gen/go/oauth2"
	"github.com/rigdev/rig/pkg/repository"
	group_service "github.com/rigdev/rig/internal/service/group"
	project_service "github.com/rigdev/rig/internal/service/project"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/hash"
	"github.com/rigdev/rig/pkg/iterator"
	"github.com/rigdev/rig/pkg/utils"
	"github.com/rigdev/rig/pkg/uuid"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service interface {
	CreateOauth2Link(ctx context.Context, p model.OauthProvider, uuid uuid.UUID, iss, sub string) (*oauth2.ProviderLink, error)
	GetOauth2Link(ctx context.Context, issuer, subject string) (uuid.UUID, *oauth2.ProviderLink, error)

	CreateUser(ctx context.Context, rm *model.RegisterMethod, initializers []*user.Update) (*user.User, error)
	GetUser(ctx context.Context, userID uuid.UUID) (*user.User, error)
	GetUserByIdentifier(ctx context.Context, id *model.UserIdentifier) (*user.User, error)
	GetPassword(ctx context.Context, userID uuid.UUID) (*model.HashingInstance, error)
	UpdateUser(ctx context.Context, userID uuid.UUID, us []*user.Update) error
	DeleteUser(ctx context.Context, userID uuid.UUID) error
	List(ctx context.Context, pagination *model.Pagination, search string) (iterator.Iterator[*model.UserEntry], uint64, error)

	GetSettings(ctx context.Context) (*settings.Settings, error)
	UpdateSettings(ctx context.Context, su []*settings.Update) error
}

type service struct {
	logger *zap.Logger
	ru     repository.User
	vcr    repository.VerificationCode
	rs     repository.Secret
	ps     project_service.Service
	gs     *group_service.Service
}

type newParams struct {
	fx.In
	Logger               *zap.Logger
	UserRepo             repository.User
	SecretRepo           repository.Secret
	ProjService          project_service.Service
	GroupService         *group_service.Service
	VerificationCodeRepo repository.VerificationCode
}

func NewService(p newParams) Service {
	return &service{
		logger: p.Logger,
		ru:     p.UserRepo,
		vcr:    p.VerificationCodeRepo,
		ps:     p.ProjService,
		gs:     p.GroupService,
		rs:     p.SecretRepo,
	}
}

func (s *service) CreateOauth2Link(ctx context.Context, p model.OauthProvider, uuid uuid.UUID, iss, sub string) (*oauth2.ProviderLink, error) {
	v := &oauth2.ProviderLink{
		Provider: p,
		Issuer:   iss,
		Subject:  sub,
	}
	return v, s.ru.CreateOauth2Link(ctx, uuid, v)
}

func (s *service) GetOauth2Link(ctx context.Context, issuer, subject string) (uuid.UUID, *oauth2.ProviderLink, error) {
	return s.ru.GetOauth2Link(ctx, issuer, subject)
}

func (s *service) CreateUser(ctx context.Context, rm *model.RegisterMethod, initializers []*user.Update) (*user.User, error) {
	set, err := s.GetSettings(ctx)
	if err != nil {
		return nil, err
	}

	if _, err := auth.GetClaims(ctx); errors.IsUnauthenticated(err) {
		// Unauthenticated access.
		if !set.GetAllowRegister() {
			s.logger.Debug("register attempt for project with register disallowed")
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	userID := uuid.New()
	u := &user.User{
		UserId: userID.String(),
		UserInfo: &model.UserInfo{
			CreatedAt: timestamppb.Now(),
		},
		Profile: &user.Profile{},
		RegisterInfo: &model.RegisterInfo{
			Method: rm,
		},
	}

	if c, err := auth.GetClaims(ctx); err == nil {
		u.RegisterInfo.CreaterId = c.GetSubject().String()
	}

	var pw *model.HashingInstance
	if err := applyUpdates(set, u, &pw, initializers); err != nil {
		return nil, err
	}
	u, err = s.ru.Create(ctx, u)
	if err != nil {
		return nil, err
	}

	if pw != nil {
		if err := s.ru.UpdatePassword(ctx, userID, pw); err != nil {
			return nil, err
		}
	}

	// Send welcome email if user - i.e. don't send mail for migration
	if set.GetSendWelcomeMail() {
		switch rm.GetMethod().(type) {
		case *model.RegisterMethod_System_,
			*model.RegisterMethod_Signup_,
			*model.RegisterMethod_OauthProvider:
			// System registration, send email
			err = s.sendWelcomeEmail(ctx, userID, u)
			if err != nil {
				s.logger.Warn("failed to send welcome email", zap.Error(err))
			}
		}
	}

	return u, nil
}

func (s *service) UpdateUser(ctx context.Context, userID uuid.UUID, us []*user.Update) error {
	set, err := s.GetSettings(ctx)
	if err != nil {
		return err
	}

	u, err := s.ru.Get(ctx, userID)
	if err != nil {
		return err
	}

	var pw *model.HashingInstance
	if err := applyUpdates(set, u, &pw, us); err != nil {
		return err
	}

	if _, err := s.ru.Update(ctx, u); err != nil {
		return err
	}

	if pw != nil {
		if err := s.ru.UpdatePassword(ctx, userID, pw); err != nil {
			return err
		}
	}

	return nil
}

func (s *service) GetUser(ctx context.Context, userID uuid.UUID) (*user.User, error) {
	user, err := s.ru.Get(ctx, userID)
	if err != nil {
		return nil, err
	}
	it, _, err := s.gs.ListGroupsForUser(ctx, userID, &model.Pagination{})
	if err != nil {
		return nil, err
	}
	defer it.Close()
	for {
		g, err := it.Next()
		if err == io.EOF {
			return user, nil
		} else if err != nil {
			return nil, err
		}
		user.GetUserInfo().GroupIds = append(user.GetUserInfo().GetGroupIds(), g.GetGroupId())
	}
}

func (s *service) GetPassword(ctx context.Context, userID uuid.UUID) (*model.HashingInstance, error) {
	return s.ru.GetPassword(ctx, userID)
}

func (s *service) GetUserByIdentifier(ctx context.Context, id *model.UserIdentifier) (*user.User, error) {
	user, err := s.ru.GetByIdentifier(ctx, id)
	if err != nil {
		return nil, err
	}
	it, _, err := s.gs.ListGroupsForUser(ctx, uuid.UUID(user.GetUserId()), &model.Pagination{})
	if err != nil {
		return nil, err
	}
	defer it.Close()
	for {
		g, err := it.Next()
		if err == io.EOF {
			return user, nil
		} else if err != nil {
			return nil, err
		}
		user.GetUserInfo().GroupIds = append(user.GetUserInfo().GetGroupIds(), g.GetGroupId())
	}
}

func (s *service) List(ctx context.Context, pagination *model.Pagination, search string) (iterator.Iterator[*model.UserEntry], uint64, error) {
	set, err := s.GetSettings(ctx)
	if err != nil {
		return nil, 0, err
	}

	uit, total, err := s.ru.List(ctx, set, pagination, search)
	if err != nil {
		return nil, 0, err
	}

	itm := iterator.Map(uit, func(u *model.UserEntry) (*model.UserEntry, error) {
		uid := uuid.UUID(u.GetUserId())

		git, _, err := s.gs.ListGroupsForUser(ctx, uid, &model.Pagination{})
		if err != nil {
			return nil, err
		}
		defer git.Close()
		for {
			g, err := git.Next()
			if err == io.EOF {
				return u, nil
			} else if err != nil {
				return nil, err
			}
			u.GroupIds = append(u.GetGroupIds(), g.GetGroupId())
		}
	})

	return itm, total, nil
}

func (s *service) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	_, err := s.ru.Delete(ctx, userID)
	if err != nil {
		return err
	}

	err = s.gs.RemoveMemberFromAll(ctx, userID)
	if err != nil {
		return err
	}

	err = s.ru.DeleteOauth2Links(ctx, userID)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) GetSettings(ctx context.Context) (*settings.Settings, error) {
	res := &settings.Settings{}
	err := s.ps.GetSettings(ctx, project_service.SettingsTypeUsers, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *service) UpdateSettings(ctx context.Context, su []*settings.Update) error {
	set, err := s.GetSettings(ctx)
	if err != nil {
		return err
	}
	err = s.applySettingsUpdates(ctx, set, su)
	if err != nil {
		return err
	}
	return s.ps.SetSettings(ctx, project_service.SettingsTypeUsers, set)
}

func (s *service) applySettingsUpdates(ctx context.Context, set *settings.Settings, su []*settings.Update) error {
	for _, up := range su {
		switch v := up.GetField().(type) {
		case *settings.Update_AllowRegister:
			set.AllowRegister = v.AllowRegister
		case *settings.Update_IsVerifiedEmailRequired:
			set.IsVerifiedEmailRequired = v.IsVerifiedEmailRequired
		case *settings.Update_IsVerifiedPhoneRequired:
			set.IsVerifiedPhoneRequired = v.IsVerifiedPhoneRequired
		case *settings.Update_AccessTokenTtl:
			set.AccessTokenTtl = v.AccessTokenTtl
		case *settings.Update_RefreshTokenTtl:
			set.RefreshTokenTtl = v.RefreshTokenTtl
		case *settings.Update_VerificationCodeTtl:
			set.VerificationCodeTtl = v.VerificationCodeTtl
		case *settings.Update_PasswordHashing:
			set.PasswordHashing = v.PasswordHashing
		case *settings.Update_LoginMechanisms_:
			set.LoginMechanisms = v.LoginMechanisms.GetLoginMechanisms()
		case *settings.Update_CallbackUrls_:
			set.GetOauthSettings().CallbackUrls = v.CallbackUrls.GetCallbackUrls()
		case *settings.Update_OauthProvider:
			switch v.OauthProvider.Provider {
			case model.OauthProvider_OAUTH_PROVIDER_FACEBOOK:
				sid, _ := uuid.Parse(set.GetOauthSettings().GetFacebook().GetSecretId())

				sID, err := s.updateSecretCreds(ctx, sid, v.OauthProvider.GetCredentials())
				if err != nil {
					return err
				}

				set.GetOauthSettings().GetFacebook().ClientId = v.OauthProvider.GetCredentials().GetPublicKey()
				set.GetOauthSettings().GetFacebook().SecretId = sID.String()
				set.GetOauthSettings().GetFacebook().AllowLogin = v.OauthProvider.GetAllowLogin()
				set.GetOauthSettings().GetFacebook().AllowRegister = v.OauthProvider.GetAllowRegister()
			case model.OauthProvider_OAUTH_PROVIDER_GITHUB:
				sid, _ := uuid.Parse(set.GetOauthSettings().GetGithub().GetSecretId())

				sID, err := s.updateSecretCreds(ctx, sid, v.OauthProvider.GetCredentials())
				if err != nil {
					return err
				}

				set.GetOauthSettings().GetGithub().ClientId = v.OauthProvider.GetCredentials().GetPublicKey()
				set.GetOauthSettings().GetGithub().SecretId = sID.String()
				set.GetOauthSettings().GetGithub().AllowLogin = v.OauthProvider.GetAllowLogin()
				set.GetOauthSettings().GetGithub().AllowRegister = v.OauthProvider.GetAllowRegister()
			case model.OauthProvider_OAUTH_PROVIDER_GOOGLE:
				sid, _ := uuid.Parse(set.GetOauthSettings().GetGoogle().GetSecretId())

				sID, err := s.updateSecretCreds(ctx, sid, v.OauthProvider.GetCredentials())
				if err != nil {
					return err
				}

				set.GetOauthSettings().GetGoogle().ClientId = v.OauthProvider.GetCredentials().GetPublicKey()
				set.GetOauthSettings().GetGoogle().SecretId = sID.String()
				set.GetOauthSettings().GetGoogle().AllowLogin = v.OauthProvider.GetAllowLogin()
				set.GetOauthSettings().GetGoogle().AllowRegister = v.OauthProvider.GetAllowRegister()
			default:
				return errors.InvalidArgumentErrorf("invalid oauth provider type '%v'", v.OauthProvider.Provider)
			}
		default:
			return errors.InvalidArgumentErrorf("invalid field type, %v", reflect.TypeOf(v))
		}
	}
	return nil
}

func (s *service) updateSecretCreds(ctx context.Context, sID uuid.UUID, creds *model.ProviderCredentials) (uuid.UUID, error) {
	credsBytes, err := proto.Marshal(creds)
	if err != nil {
		return uuid.Nil, err
	}

	if sID == uuid.Nil {
		secretId := uuid.New()
		s.rs.Create(ctx, secretId, credsBytes)
		return secretId, nil
	} else {
		s.rs.Update(ctx, sID, credsBytes)
		return sID, nil
	}
}

func applyUpdates(set *settings.Settings, u *user.User, pw **model.HashingInstance, us []*user.Update) error {
	for _, up := range us {
		switch v := up.GetField().(type) {
		case *user.Update_Email:
			if _, err := mail.ParseAddress(v.Email); err != nil {
				return errors.InvalidArgumentErrorf("invalid email address")
			}
			u.UserInfo.Email = v.Email
			u.IsEmailVerified = false
		case *user.Update_Username:
			u.UserInfo.Username = v.Username
		case *user.Update_PhoneNumber:
			p, err := phonenumbers.Parse(v.PhoneNumber, "")
			if err != nil {
				return errors.InvalidArgumentErrorf("invalid phone number; %v", err)
			}

			if !phonenumbers.IsValidNumber(p) {
				return errors.InvalidArgumentErrorf("invalid phone number for country")
			}

			u.UserInfo.PhoneNumber = phonenumbers.Format(p, phonenumbers.E164)
			u.IsPhoneVerified = false
		case *user.Update_IsEmailVerified:
			u.IsEmailVerified = v.IsEmailVerified
		case *user.Update_Password:
			if err := utils.ValidatePassword(v.Password); err != nil {
				return err
			}

			h := hash.New(set.GetPasswordHashing())
			npw, err := h.Generate(v.Password)
			if err != nil {
				return err
			}
			*pw = npw
		case *user.Update_ResetSessions_:
			u.NewSessionsSince = timestamppb.Now()
		case *user.Update_Profile:
			u.Profile = v.Profile
		case *user.Update_SetMetadata:
			if u.Metadata == nil {
				u.Metadata = map[string][]byte{}
			}
			u.Metadata[v.SetMetadata.GetKey()] = v.SetMetadata.GetValue()
		case *user.Update_DeleteMetadataKey:
			delete(u.Metadata, v.DeleteMetadataKey)
		case *user.Update_HashedPassword:
			*pw = v.HashedPassword
		default:
			return errors.InvalidArgumentErrorf("invalid user update type '%v'", reflect.TypeOf(up.GetField()))
		}
	}
	return nil
}
