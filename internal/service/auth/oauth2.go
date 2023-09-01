package auth

import (
	"context"
	"fmt"
	"net/url"

	"github.com/golang/protobuf/proto"
	"go.uber.org/zap"

	"github.com/rigdev/rig-go-api/api/v1/authentication"
	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig-go-api/api/v1/user/settings"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
)

func (s *Service) GetOauth2Providers(ctx context.Context, redirect string) ([]*authentication.OauthProvider, error) {
	s.logger.Debug("GetOauth2Providers", zap.String("redirect", redirect))
	pID, err := auth.GetProjectID(ctx)
	if err != nil {
		return nil, err
	}
	s.logger.Debug("GetOauth2Providers", zap.String("project_id", pID.String()))

	us, err := s.us.GetSettings(ctx)
	if err != nil {
		return nil, err
	}
	oauthproviders := []*authentication.OauthProvider{}
	rigRedirect, err := url.JoinPath(s.cfg.Management.PublicURL, "oauth/callback")
	if err != nil {
		return nil, fmt.Errorf("could not build callback URL: %w", err)
	}

	if us.GetOauthSettings().GetGithub().GetAllowLogin() {
		sid := uuid.UUID(us.GetOauthSettings().GetGithub().GetSecretId())

		secret, err := s.rs.Get(ctx, sid)
		if err != nil {
			return nil, err
		}

		creds := model.ProviderCredentials{}

		err = proto.Unmarshal(secret, &creds)
		if err != nil {
			return nil, err
		}

		url, err := s.oauth2.Github.RedirectUrl(rigRedirect, redirect, pID, &creds)
		if err != nil {
			return nil, err
		}
		oauthproviders = append(oauthproviders, &authentication.OauthProvider{
			Name:        "Github",
			ProviderUrl: url,
		})

		s.logger.Debug("GetOauth2Providers", zap.String("github url", url))
	}

	if us.GetOauthSettings().GetGoogle().GetAllowLogin() {
		sid := uuid.UUID(us.GetOauthSettings().GetGoogle().GetSecretId())

		secret, err := s.rs.Get(ctx, sid)
		if err != nil {
			return nil, err
		}

		creds := model.ProviderCredentials{}

		err = proto.Unmarshal(secret, &creds)
		if err != nil {
			return nil, err
		}
		url, err := s.oauth2.Google.RedirectUrl(rigRedirect, redirect, pID, &creds)
		if err != nil {
			return nil, err
		}
		oauthproviders = append(oauthproviders, &authentication.OauthProvider{
			Name:        "Google",
			ProviderUrl: url,
		})
		s.logger.Debug("GetOauth2Providers", zap.String("google url", url))
	}

	if us.GetOauthSettings().GetFacebook().GetAllowLogin() {
		sid := uuid.UUID(us.GetOauthSettings().GetFacebook().GetSecretId())

		secret, err := s.rs.Get(ctx, sid)
		if err != nil {
			return nil, err
		}

		creds := model.ProviderCredentials{}

		err = proto.Unmarshal(secret, &creds)
		if err != nil {
			return nil, err
		}

		url, err := s.oauth2.Facebook.RedirectUrl(rigRedirect, redirect, pID, &creds)
		if err != nil {
			return nil, err
		}
		oauthproviders = append(oauthproviders, &authentication.OauthProvider{
			Name:        "Facebook",
			ProviderUrl: url,
		})

		s.logger.Debug("GetOauth2Providers", zap.String("facebook url", url))
	}

	return oauthproviders, nil
}

func (s *Service) LoginOauth2(ctx context.Context, claims *auth.Oauth2Claims, provider *settings.OauthProviderSettings) (*authentication.Token, error) {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return nil, err
	}
	us, err := s.us.GetSettings(ctx)
	if err != nil {
		return nil, err
	}
	userID, _, err := s.us.GetOauth2Link(ctx, claims.Iss, claims.Sub)
	if errors.IsNotFound(err) {
		u, err := s.us.GetUserByIdentifier(ctx, &model.UserIdentifier{Identifier: &model.UserIdentifier_Email{Email: claims.Email}})
		if errors.IsNotFound(err) {
			// Check if allowed register on oauthprovider
			if !provider.GetAllowRegister() {
				return nil, errors.PermissionDeniedErrorf("Register with oauth not allowed")
			}
			// Create user.
			if u, err = s.us.CreateUser(ctx, &model.RegisterMethod{Method: &model.RegisterMethod_OauthProvider{}}, []*user.Update{
				{
					Field: &user.Update_Email{
						Email: claims.Email,
					},
				},
				{
					Field: &user.Update_Profile{
						Profile: &user.Profile{
							FirstName: claims.FirstName,
							LastName:  claims.LastName,
						},
					},
				},
			}); err != nil {
				return nil, err
			}
		} else if err != nil {
			return nil, err
		}

		userID = uuid.UUID(u.GetUserId())

		if _, err := s.us.CreateOauth2Link(ctx, model.OauthProvider_OAUTH_PROVIDER_UNSPECIFIED, userID, claims.Iss, claims.Sub); err != nil {
			return nil, err
		}
	}

	sessionID, ss, err := s.newSession(ctx, userID, &user.AuthMethod{Method: &user.AuthMethod_OauthProvider{}})
	if err != nil {
		return nil, err
	}

	return s.generateToken(ctx, sessionID, ss, projectID, userID, auth.SubjectTypeUser, nil, us)
}
