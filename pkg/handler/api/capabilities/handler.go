package capabilities

import (
	"bytes"
	"context"

	connect "connectrpc.com/connect"
	"gopkg.in/yaml.v3"

	"github.com/rigdev/rig-go-api/operator/api/v1/capabilities"
	"github.com/rigdev/rig-go-api/operator/api/v1/capabilities/capabilitiesconnect"
	svccapabilities "github.com/rigdev/rig/pkg/service/capabilities"
	"github.com/rigdev/rig/pkg/service/config"
)

func NewHandler(capabilities svccapabilities.Service, cfg config.Service) capabilitiesconnect.ServiceHandler {
	return &handler{
		capabilities: capabilities,
		cfg:          cfg,
	}
}

type handler struct {
	capabilities svccapabilities.Service
	cfg          config.Service
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

func (h *handler) GetConfig(ctx context.Context, _ *connect.Request[capabilities.GetConfigRequest]) (*connect.Response[capabilities.GetConfigResponse], error) {
	cfg := h.cfg.Operator()
	buffer := &bytes.Buffer{}
	encoder := yaml.NewEncoder(buffer)
	if err := encoder.Encode(cfg); err != nil {
		return nil, err
	}

	return connect.NewResponse(&capabilities.GetConfigResponse{
		Yaml: buffer.String(),
	}), nil
}
