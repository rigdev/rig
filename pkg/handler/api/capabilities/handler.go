package capabilities

import (
	"context"

	connect "connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/operator/api/v1/capabilities"
	"github.com/rigdev/rig-go-api/operator/api/v1/capabilities/capabilitiesconnect"
	"github.com/rigdev/rig/pkg/obj"
	svccapabilities "github.com/rigdev/rig/pkg/service/capabilities"
	"github.com/rigdev/rig/pkg/service/config"
	"k8s.io/apimachinery/pkg/runtime"
)

func NewHandler(
	capabilities svccapabilities.Service,
	cfg config.Service,
	scheme *runtime.Scheme,
) capabilitiesconnect.ServiceHandler {
	return &handler{
		capabilities: capabilities,
		cfg:          cfg,
		scheme:       scheme,
	}
}

type handler struct {
	capabilities svccapabilities.Service
	cfg          config.Service
	scheme       *runtime.Scheme
}

// Get implements capabilitiesconnect.ServiceClient.
func (h *handler) Get(
	ctx context.Context,
	_ *connect.Request[capabilities.GetRequest],
) (*connect.Response[capabilities.GetResponse], error) {
	res, err := h.capabilities.Get(ctx)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(res), nil
}

func (h *handler) GetConfig(
	_ context.Context,
	_ *connect.Request[capabilities.GetConfigRequest],
) (*connect.Response[capabilities.GetConfigResponse], error) {
	cfg := h.cfg.Operator()
	bytes, err := obj.Encode(cfg, h.scheme)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&capabilities.GetConfigResponse{
		Yaml: string(bytes),
	}), nil
}

func (h *handler) GetPlugins(
	_ context.Context,
	_ *connect.Request[capabilities.GetPluginsRequest],
) (*connect.Response[capabilities.GetPluginsResponse], error) {
	return connect.NewResponse(h.capabilities.GetMods()), nil
}
