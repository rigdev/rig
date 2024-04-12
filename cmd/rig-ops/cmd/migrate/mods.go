package migrate

import (
	"context"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/operator/api/v1/capabilities"
	"github.com/rigdev/rig/cmd/rig-ops/cmd/base"
	"github.com/rigdev/rig/pkg/controller/mod"
)

// Get mods from the operator config that matches the name of the capsule and the namespace
func (c *Cmd) getMods(ctx context.Context, migration *Migration) error {
	cfg, err := base.GetOperatorConfig(ctx, c.OperatorClient, c.Scheme)
	if err != nil {
		return err
	}
	resp, err := c.OperatorClient.Capabilities.GetPlugins(ctx, connect.NewRequest(&capabilities.GetPluginsRequest{}))
	if err != nil {
		return err
	}

	var mods []string
	for _, step := range cfg.Pipeline.Steps {
		matcher, err := mod.NewMatcher(step.Namespaces, step.Capsules, step.Selector, false)
		if err != nil {
			return err
		}
		if !matcher.Match(migration.capsule.Namespace, migration.capsule.Name, migration.capsule.Annotations) {
			continue
		}

		// Mod matches capsule
		for _, p := range step.Plugins {
			for _, mod := range resp.Msg.Plugins {
				switch v := mod.GetPlugin().(type) {
				case *capabilities.GetPluginsResponse_Plugin_Builtin:
					if v.Builtin.GetName() == p.Name {
						mods = append(mods, p.Name)
					}
				case *capabilities.GetPluginsResponse_Plugin_ThirdParty:
					if v.ThirdParty.Name == p.Name {
						mods = append(mods, p.Name)
					}
				}
			}
		}
	}

	migration.mods = mods

	return nil
}
