package capabilities

import (
	"context"

	connect "connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/operator/api/v1/capabilities"
	"github.com/rigdev/rig-go-api/operator/api/v1/capabilities/capabilitiesconnect"
	"github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/obj"
	svccapabilities "github.com/rigdev/rig/pkg/service/capabilities"
	"k8s.io/apimachinery/pkg/runtime"
)

func NewHandler(
	capabilities svccapabilities.Service,
	cfg *v1alpha1.OperatorConfig,
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
	scheme       *runtime.Scheme
	cfg          *v1alpha1.OperatorConfig
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
	bytes, err := obj.Encode(h.cfg, h.scheme)
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
	return connect.NewResponse(h.capabilities.GetPlugins()), nil
}
