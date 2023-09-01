package gateway

import (
	"fmt"

	"github.com/rigdev/rig/internal/client/docker"
	"github.com/rigdev/rig/internal/client/k8s"
	"github.com/rigdev/rig/internal/client/minio"
	"github.com/rigdev/rig/internal/gateway/cluster"
	"github.com/rigdev/rig/internal/gateway/storage"
	"github.com/rigdev/rig/internal/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module(
	"gateway",
	fx.Provide(
		NewCluster,
		NewStorage,
	),
)

type clusterParams struct {
	fx.In

	Cfg    config.Config
	Logger *zap.Logger

	DockerClient *docker.Client `optional:"true"`
	K8SClient    *k8s.Client    `optional:"true"`
}

type storageParams struct {
	fx.In

	MinioClient *minio.Storage
	Logger      *zap.Logger
}

func NewCluster(p clusterParams) (cluster.Gateway, error) {
	switch p.Cfg.Cluster.Type {
	case "docker":
		if p.DockerClient == nil {
			return nil, fmt.Errorf("no docker client provided")
		}
		return p.DockerClient, nil
	case "k8s":
		if p.K8SClient == nil {
			return nil, fmt.Errorf("no k8s client provided")
		}
		return p.K8SClient, nil
	default:
		return nil, fmt.Errorf("invalid cluster gateway '%v'", p.Cfg.Cluster.Type)
	}
}

func NewStorage(p storageParams) (storage.Gateway, error) {
	if p.MinioClient == nil {
		return nil, fmt.Errorf("no minio client provided")
	}
	return p.MinioClient, nil
}
