package auth

import (
	"context"
	"testing"

	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig-go-api/api/v1/user/settings"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/gen/go/oauth2"
	"github.com/rigdev/rig/internal/config"
	"github.com/rigdev/rig/internal/repository"
	"github.com/rigdev/rig/internal/service/project"
	user_serv "github.com/rigdev/rig/internal/service/user"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_Login_Oauth2_User_Not_Exists_Success(t *testing.T) {
	us := user_serv.NewMockService(t)
	ps := project.NewMockService(t)
	sr := repository.NewMockSession(t)
	projectID := uuid.New()
	ctx := auth.WithProjectID(context.Background(), projectID)

	s, err := NewService(newServiceParams{
		Config:      config.Config{Auth: config.Auth{JWT: config.AuthJWT{Secret: "jwtsecret"}}},
		UserService: us,
		ProjService: ps,
		SessionRepo: sr,
	})
	require.NoError(t, err)

	us.EXPECT().GetSettings(ctx).Return(&settings.Settings{}, nil)

	claims := auth.Oauth2Claims{
		Iss:       "google",
		Sub:       "1234567890",
		Email:     "test@rig.dev",
		FirstName: "Tester",
		LastName:  "Testingson",
	}

	us.EXPECT().GetOauth2Link(ctx, claims.Iss, claims.Sub).Return(uuid.Nil, nil, errors.NotFoundErrorf("user not found"))

	us.EXPECT().GetUserByIdentifier(ctx, &model.UserIdentifier{
		Identifier: &model.UserIdentifier_Email{
			Email: claims.Email,
		},
	}).Return(nil, errors.NotFoundErrorf("user not found"))

	userID := uuid.New()
	us.EXPECT().CreateUser(ctx, &model.RegisterMethod{Method: &model.RegisterMethod_OauthProvider{}}, []*user.Update{
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
	}).Return(&user.User{
		UserId: userID.String(),
	}, nil)

	us.EXPECT().CreateOauth2Link(ctx, model.OauthProvider_OAUTH_PROVIDER_UNSPECIFIED, userID, claims.Iss, claims.Sub).Return(&oauth2.ProviderLink{}, nil)

	sr.EXPECT().Create(ctx, userID, mock.Anything, mock.Anything).Return(nil)
	sr.EXPECT().Update(ctx, userID, mock.Anything, mock.Anything).Return(nil)

	prov := &settings.OauthProviderSettings{
		AllowRegister: true,
	}

	_, err = s.LoginOauth2(ctx, &claims, prov)
	require.NoError(t, err)
}
