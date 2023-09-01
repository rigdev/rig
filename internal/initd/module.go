package initd

import (
	"go.uber.org/fx"
)

var Module = fx.Module(
	"service",
	fx.Provide(New),
)
