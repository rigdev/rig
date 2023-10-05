package service

import (
	"github.com/rigdev/rig/internal/service/auth"
	"github.com/rigdev/rig/internal/service/capsule"
	"github.com/rigdev/rig/internal/service/cluster"
	"github.com/rigdev/rig/internal/service/group"
	"github.com/rigdev/rig/internal/service/metrics"
	"github.com/rigdev/rig/internal/service/project"
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
		metrics.NewService,
		cluster.NewService,
	),
)
