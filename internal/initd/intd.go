package initd

import (
	"github.com/rigdev/rig/internal/config"
	user_service "github.com/rigdev/rig/internal/service/user"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Service struct {
	logger *zap.Logger
	us     user_service.Service
	cfg    config.Config
}

type newParams struct {
	fx.In
	Logger      *zap.Logger
	UserService user_service.Service
	Config      config.Config
}

func New(p newParams) *Service {
	s := &Service{
		logger: p.Logger,
		us:     p.UserService,
		cfg:    p.Config,
	}
	return s
}
