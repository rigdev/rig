package config

import (
	"github.com/rigdev/rig/cmd/rig/cmd/base"
)

func ConfigUseContext(args []string, cfg *base.Config) error {
	if len(args) > 0 {
		return base.UseContext(cfg, args[0])
	}

	return base.SelectContext(cfg)
}
