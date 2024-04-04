package scope

import (
	"github.com/rigdev/rig/cmd/rig/cmd/cmdconfig"
)

type Interactive bool

// The Scope that the command is running in. The scope is build from a mix og the cfg and the provided flags
type Scope interface {
	GetCfg() *cmdconfig.Config
	GetCurrentContext() *cmdconfig.Context
	IsInteractive() bool
}

type scope struct {
	cfg         *cmdconfig.Config
	context     *cmdconfig.Context
	Interactive bool
}

func NewScope(cfg *cmdconfig.Config, ctx *cmdconfig.Context, interactive Interactive) Scope {
	return &scope{
		cfg:         cfg,
		context:     ctx,
		Interactive: bool(interactive),
	}
}

func (s *scope) GetCurrentContext() *cmdconfig.Context {
	return s.context
}

func (s *scope) IsInteractive() bool {
	return s.Interactive
}

func (s *scope) GetCfg() *cmdconfig.Config {
	return s.cfg
}
