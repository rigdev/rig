package auth

import (
	"context"
	"testing"

	api_project "github.com/rigdev/rig-go-api/api/v1/project"
	api_user "github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig-go-api/api/v1/user/settings"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/internal/config"
	"github.com/rigdev/rig/pkg/repository"
	"github.com/rigdev/rig/internal/service/project"
	"github.com/rigdev/rig/internal/service/user"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/hash"
	"github.com/rigdev/rig/pkg/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_Login_UserPassword_Success(t *testing.T) {
	us := user.NewMockService(t)
	ps := project.NewMockService(t)
	sr := repository.NewMockSession(t)
	projectID := uuid.New().String()
	ctx := auth.WithProjectID(context.Background(), projectID)

	s, err := NewService(newServiceParams{
		Config:      config.Config{Auth: config.Auth{JWT: config.AuthJWT{Secret: "jwtsecret"}}},
		UserService: us,
		ProjService: ps,
		SessionRepo: sr,
	})
	require.NoError(t, err)

	id := &model.UserIdentifier{
		Identifier: &model.UserIdentifier_Email{
			Email: "test@rig.dev",
		},
	}

	userID := uuid.New()
	user := &api_user.User{
		UserId: userID.String(),
	}

	us.EXPECT().GetUserByIdentifier(ctx, id).Return(user, nil)

	ps.EXPECT().GetProject(ctx).Return(&api_project.Project{}, nil)

	us.EXPECT().GetSettings(ctx).Return(&settings.Settings{}, nil)

	h := hash.New(&model.HashingConfig{
		Method: &model.HashingConfig_Bcrypt{
			Bcrypt: hash.DefaultBcrypt,
		},
	})
	pwh, err := h.Generate("secretpassword")
	require.NoError(t, err)

	us.EXPECT().GetPassword(ctx, userID).Return(pwh, nil)

	sr.EXPECT().Create(ctx, userID, mock.Anything, mock.Anything).Return(nil)
	sr.EXPECT().Update(ctx, userID, mock.Anything, mock.Anything).Return(nil)

	_, _, _, err = s.LoginUserPassword(ctx, id, "secretpassword")
	require.NoError(t, err)
}
