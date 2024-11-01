package cluster

import (
	"context"

	"connectrpc.com/connect"
	api_cluster "github.com/rigdev/rig-go-api/operator/api/v1/cluster"
	"github.com/rigdev/rig-go-api/operator/api/v1/cluster/clusterconnect"
	"github.com/rigdev/rig/pkg/service/cluster"
)

func NewHandler(
	cluster cluster.Service,
) clusterconnect.ServiceHandler {
	return &handler{
		cluster: cluster,
	}
}

type handler struct {
	cluster cluster.Service
}

func (h *handler) GetNodes(
	ctx context.Context, _ *connect.Request[api_cluster.GetNodesRequest],
) (*connect.Response[api_cluster.GetNodesResponse], error) {
	nodes, err := h.cluster.GetNodes(ctx)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&api_cluster.GetNodesResponse{
		Nodes: nodes,
	}), nil
}

func (h *handler) GetNodePods(
	ctx context.Context, req *connect.Request[api_cluster.GetNodePodsRequest],
) (*connect.Response[api_cluster.GetNodePodsResponse], error) {
	pods, err := h.cluster.GetNodePods(ctx, req.Msg.GetNodeName())
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&api_cluster.GetNodePodsResponse{
		Pods: pods,
	}), nil
}
