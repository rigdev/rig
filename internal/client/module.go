package client

import (
	"github.com/rigdev/rig/internal/client/docker"
	"github.com/rigdev/rig/internal/client/k8s"
	"github.com/rigdev/rig/internal/client/minio"
	"github.com/rigdev/rig/internal/client/mongo"
	"github.com/rigdev/rig/internal/client/postgres"
	"github.com/rigdev/rig/internal/client/segment"
	"github.com/rigdev/rig/internal/config"
	"go.uber.org/fx"
)

func GetModule(cfg config.Config) fx.Option {
	var opts []fx.Option
	if cfg.Client.Mongo.Host != "" {
		opts = append(opts, fx.Provide(mongo.New))
	}
	if cfg.Client.Postgres.Host != "" {
		opts = append(opts, fx.Provide(postgres.New))
	}
	switch cfg.Cluster.Type {
	case config.ClusterTypeDocker:
		opts = append(opts, fx.Provide(docker.New))
	case config.ClusterTypeKubernetes,
		config.ClusterTypeKubernetesNative:
		opts = append(opts, fx.Provide(k8s.New))
	}
	return fx.Module(
		"clients",
		fx.Provide(
			minio.NewDefault,
			segment.New,
		),
		fx.Options(
			opts...,
		),
	)
}
