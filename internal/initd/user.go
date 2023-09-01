package initd

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig-go-api/model"
	auth_service "github.com/rigdev/rig/internal/service/auth"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
)

const (
	InitUserUsername = "rig-init-user"
)

func (s *Service) InitUser() error {
	ctx := auth.WithProjectID(context.Background(), auth.RigProjectID)
	ctx = auth.WithClaims(ctx, auth_service.ProjectClaims{
		UseProjectID: auth.RigProjectID,
	})
	if s.cfg.Init.Root.Email != "" && s.cfg.Init.Root.Password != "" {
		_, err := s.us.GetUserByIdentifier(ctx, &model.UserIdentifier{Identifier: &model.UserIdentifier_Username{Username: InitUserUsername}})
		if err == nil {
			s.logger.Info("init user already exist")
			return nil
		} else {
			if !errors.IsNotFound(err) {
				return err
			}
		}
		s.logger.Info("init user does not exist... creating...")
		if _, err := s.us.CreateUser(ctx, &model.RegisterMethod{Method: &model.RegisterMethod_System_{}}, []*user.Update{
			{Field: &user.Update_Email{Email: s.cfg.Init.Root.Email}},
			{Field: &user.Update_Password{Password: s.cfg.Init.Root.Password}},
			{Field: &user.Update_Username{Username: InitUserUsername}},
		}); err != nil {
			return err
		}
		s.logger.Info("created init user...")
	}
	return nil
}
