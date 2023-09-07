package config

import (
	"github.com/rigdev/rig/cmd/rig/cmd/cmd_config"
)

func ConfigUseContext(args []string, cfg *cmd_config.Config) error {
	if len(args) > 0 {
		return cmd_config.UseContext(cfg, args[0])
	}

	return cmd_config.SelectContext(cfg)
}
