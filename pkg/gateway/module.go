package gateway

import (
	"fmt"

	"github.com/rigdev/rig/pkg/client/docker"
	"github.com/rigdev/rig/pkg/client/k8s"
	"github.com/rigdev/rig/pkg/client/minio"
	"github.com/rigdev/rig/internal/config"
	"github.com/rigdev/rig/pkg/gateway/cluster"
	"github.com/rigdev/rig/pkg/gateway/storage"
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

func NewCluster(p clusterParams) (cluster.Gateway, cluster.ConfigGateway, cluster.StatusGateway, error) {
	switch p.Cfg.Cluster.Type {
	case config.ClusterTypeDocker:
		if p.DockerClient == nil {
			return nil, nil, nil, fmt.Errorf("no docker client provided")
		}
		return p.DockerClient, p.DockerClient, p.DockerClient, nil
	case config.ClusterTypeKubernetes:
		if p.K8SClient == nil {
			return nil, nil, nil, fmt.Errorf("no k8s client provided")
		}
		return p.K8SClient, p.K8SClient, nil, nil
	case config.ClusterTypeKubernetesNative:
		if p.K8SClient == nil {
			return nil, nil, nil, fmt.Errorf("no k8s client provided")
		}
		return p.K8SClient, p.K8SClient.ConfigGateway(), nil, nil
	default:
		return nil, nil, nil, fmt.Errorf("invalid cluster gateway '%v'", p.Cfg.Cluster.Type)
	}
}

func NewStorage(p storageParams) (storage.Gateway, error) {
	if p.MinioClient == nil {
		return nil, fmt.Errorf("no minio client provided")
	}
	return p.MinioClient, nil
}
