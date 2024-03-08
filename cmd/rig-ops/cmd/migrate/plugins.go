package migrate

import (
	"context"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/operator/api/v1/capabilities"
	"github.com/rigdev/rig/cmd/rig-ops/cmd/base"
	"k8s.io/apimachinery/pkg/runtime"
)

// Get plugins from the operator config that matches all capsules and namespaces and is installed in the cluster
func getPlugins(ctx context.Context, operatorClient *base.OperatorClient, scheme *runtime.Scheme) ([]string, error) {
	cfg, err := base.GetOperatorConfig(ctx, operatorClient, scheme)
	if err != nil {
		return nil, err
	}
	resp, err := operatorClient.Capabilities.GetPlugins(ctx, connect.NewRequest(&capabilities.GetPluginsRequest{}))
	if err != nil {
		return nil, err
	}

	var plugins []string
	for _, step := range cfg.Pipeline.Steps {
		if step.Capsules != nil && step.Namespaces != nil && step.Selector.Size() > 0 {
			continue
		}

		// Plugin matches all capsules
		for _, p := range step.Plugins {
			for _, plugin := range resp.Msg.Plugins {
				switch v := plugin.GetPlugin().(type) {
				case *capabilities.GetPluginsResponse_Plugin_Builtin:
					if v.Builtin.GetType() == p.Type {
						plugins = append(plugins, p.Type)
					}
				case *capabilities.GetPluginsResponse_Plugin_ThirdParty:
					if v.ThirdParty.GetType() == p.Type {
						plugins = append(plugins, p.Type)
					}
				}
			}
		}
	}

	return plugins, nil
}
