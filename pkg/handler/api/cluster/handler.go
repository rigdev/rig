package cluster

import (
	"context"
	"net/http"

	connect "github.com/bufbuild/connect-go"
	cluster_api "github.com/rigdev/rig-go-api/api/v1/cluster"
	"github.com/rigdev/rig-go-api/api/v1/cluster/clusterconnect"
	"github.com/rigdev/rig/internal/service/cluster"
)

type Handler struct {
	cluster *cluster.Service
}

// New returns an impementation of the proto cluster.Service interface.
func New(cluster *cluster.Service) *Handler {
	return &Handler{
		cluster: cluster,
	}
}

func (h *Handler) ServiceName() string {
	return clusterconnect.ServiceName
}

func (h *Handler) Build(opts ...connect.HandlerOption) (string, http.Handler) {
	return clusterconnect.NewServiceHandler(h, opts...)
}

func (h *Handler) GetConfig(ctx context.Context, req *connect.Request[cluster_api.GetConfigRequest]) (*connect.Response[cluster_api.GetConfigResponse], error) {
	resp, err := h.cluster.GetConfig(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(resp), nil
}
