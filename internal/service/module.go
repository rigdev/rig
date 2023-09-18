package service

import (
	"github.com/rigdev/rig/internal/service/auth"
	"github.com/rigdev/rig/internal/service/capsule"
	"github.com/rigdev/rig/internal/service/database"
	"github.com/rigdev/rig/internal/service/group"
	"github.com/rigdev/rig/internal/service/metrics"
	"github.com/rigdev/rig/internal/service/operator"
	"github.com/rigdev/rig/internal/service/project"
	"github.com/rigdev/rig/internal/service/storage"
	"github.com/rigdev/rig/internal/service/user"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"service",
	fx.Provide(
		capsule.NewService,
		user.NewService,
		auth.NewService,
		project.NewService,
		group.NewService,
		database.NewService,
		storage.NewService,
		metrics.NewService,
		operator.New,
	),
)
