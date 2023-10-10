package capabilities

import (
	"context"

	connect "github.com/bufbuild/connect-go"

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
func (h *handler) Get(ctx context.Context, req *connect.Request[capabilities.GetRequest]) (*connect.Response[capabilities.GetResponse], error) {
	res, err := h.capabilities.Get(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(res), nil
}
