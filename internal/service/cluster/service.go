package cluster

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/cluster"
	"github.com/rigdev/rig/internal/config"
	"go.uber.org/zap"
)

type Service struct {
	logger *zap.Logger
	cfg    config.Config
}

func NewService(cfg config.Config, logger *zap.Logger) *Service {
	s := &Service{
		cfg:    cfg,
		logger: logger,
	}

	return s
}

func (s *Service) GetConfig(ctx context.Context, req *cluster.GetConfigRequest) (*cluster.GetConfigResponse, error) {
	resp := &cluster.GetConfigResponse{
		ClusterType: typeToProto(s.cfg.Cluster.Type),
	}
	s.setDevRegistry(resp)
	return resp, nil
}

func typeToProto(c config.ClusterType) cluster.ClusterType {
	switch c {
	case config.ClusterTypeDocker:
		return cluster.ClusterType_CLUSTER_TYPE_DOCKER
	case config.ClusterTypeKubernetes:
		return cluster.ClusterType_CLUSTER_TYPE_KUBERNETES
	default:
		return cluster.ClusterType_CLUSTER_TYPE_UNSPECIFIED
	}
}

func (s *Service) setDevRegistry(resp *cluster.GetConfigResponse) {
	if s.cfg.Cluster.Type == config.ClusterTypeDocker {
		resp.DevRegistry = &cluster.GetConfigResponse_Docker{
			Docker: &cluster.DockerDaemon{},
		}
		return
	}

	registry := s.cfg.Cluster.DevRegistry
	if registry.Host == "" {
		return
	}

	resp.DevRegistry = &cluster.GetConfigResponse_Registry{
		Registry: &cluster.Registry{
			Host: registry.Host,
		},
	}
}
