package capabilities

import (
	"context"

	connect "connectrpc.com/connect"

	"github.com/rigdev/rig-go-api/operator/api/v1/capabilities"
	"github.com/rigdev/rig-go-api/operator/api/v1/capabilities/capabilitiesconnect"
	svccapabilities "github.com/rigdev/rig/pkg/service/capabilities"
)

func NewHandler(capabilities svccapabilities.Service) capabilitiesconnect.ServiceHandler {
	return &handler{capabilities: capabilities}
}

type handler struct {
	capabilities svccapabilities.Service
}

// Get implements capabilitiesconnect.ServiceClient.
func (h *handler) Get(
	_ context.Context,
	_ *connect.Request[capabilities.GetRequest],
) (*connect.Response[capabilities.GetResponse], error) {
	res, err := h.capabilities.Get()
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(res), nil
}
