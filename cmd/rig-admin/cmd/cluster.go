package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/internal/config"
	"github.com/rigdev/rig/internal/gateway/cluster"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func init() {
	cluster := &cobra.Command{
		Use: "cluster",
	}

	bootstrap := &cobra.Command{
		Use:  "bootstrap",
		RunE: register(ClusterBootstrap),
	}
	cluster.AddCommand(bootstrap)

	createMinio := &cobra.Command{
		Use:  "create-minio",
		RunE: register(ClusterCreateMinio),
	}
	cluster.AddCommand(createMinio)

	rootCmd.AddCommand(cluster)
}

func ClusterBootstrap(ctx context.Context, cmd *cobra.Command, cg cluster.Gateway, cfg config.Config, logger *zap.Logger) error {
	if err := ClusterCreateMinio(ctx, cmd, cg, cfg, logger); err != nil {
		return err
	}

	if err := ClusterCreateMongoDB(ctx, cmd, cg, cfg, logger); err != nil {
		return err
	}

	if err := ClusterCreateRig(ctx, cmd, cg, cfg, logger); err != nil {
		return err
	}

	return nil
}

func ClusterCreateRig(ctx context.Context, cmd *cobra.Command, cg cluster.Gateway, cfg config.Config, logger *zap.Logger) error {
	cc := &cluster.Capsule{
		Image: "rig",
		Network: &capsule.Network{
			Interfaces: []*capsule.Interface{
				{
					Name: "default",
					Port: 4747,
					Public: &capsule.PublicInterface{
						Enabled: true,
						Method: &capsule.RoutingMethod{
							Kind: &capsule.RoutingMethod_LoadBalancer_{
								LoadBalancer: &capsule.RoutingMethod_LoadBalancer{
									Port: 4747,
								},
							},
						},
					},
				},
			},
		},
		Replicas: 1,
		Volumes:  map[string]string{},
	}

	switch cfg.Cluster.Type {
	case "docker":
		cc.Volumes["/var/run/docker.sock"] = "/var/run/docker.sock"
	}

	if err := cg.UpsertCapsule(ctx, "rig", cc); err != nil {
		return err
	}

	logger.Info("created rig")

	return nil
}

func ClusterCreateMongoDB(ctx context.Context, cmd *cobra.Command, cg cluster.Gateway, cfg config.Config, logger *zap.Logger) error {
	vs := map[string]string{
		"mongodb-data": "/data/db",
	}

	cc := &cluster.Capsule{
		Image:    "mongo:6.0",
		Replicas: 1,
		Volumes:  vs,
	}

	if err := cg.UpsertCapsule(ctx, "mongodb", cc); err != nil {
		return err
	}

	logger.Info("created mongodb")

	return nil
}

func ClusterCreateMinio(ctx context.Context, cmd *cobra.Command, cg cluster.Gateway, cfg config.Config, logger *zap.Logger) error {
	replicas := 3
	data := 2

	vs := map[string]string{}
	for i := 0; i < data; i++ {
		vs[fmt.Sprint("minio-data-", i)] = fmt.Sprint("/data-", i)
	}

	cc := &cluster.Capsule{
		Image:    "quay.io/minio/minio",
		Replicas: uint32(replicas),
		Volumes:  vs,
		ContainerSettings: &capsule.ContainerSettings{
			EnvironmentVariables: map[string]string{
				"MINIO_ROOT_USER":     cfg.Client.Minio.AccessKeyID,
				"MINIO_ROOT_PASSWORD": cfg.Client.Minio.SecretAccessKey,
			},
		},
	}

	switch cfg.Cluster.Type {
	case config.ClusterTypeDocker:
		cc.ContainerSettings.Command = "server"
		cc.ContainerSettings.Args = []string{
			"--console-address",
			":9001",
			fmt.Sprintf("http://minio-instance-{0...%d}/data-{0...%d}", replicas-1, data-1),
		}
	case config.ClusterTypeKubernetes:
		projectID, err := auth.GetProjectID(ctx)
		if err != nil {
			return err
		}
		ns := projectID.String()

		cc.ContainerSettings.Args = []string{
			"server",
			"--console-address",
			":9001",
			fmt.Sprintf("http://minio-{0...%d}.minio-headless.%s.svc.cluster.local/data-{0...%d}", replicas-1, ns, data-1),
		}
	default:
		return errors.New("unsupported cluster type")
	}

	if err := cg.UpsertCapsule(ctx, "minio", cc); err != nil {
		return err
	}

	logger.Info("created minio")

	minioClient, err := minio.New(cfg.Client.Minio.Host, &minio.Options{
		Creds:  credentials.NewStaticV2(cfg.Client.Minio.AccessKeyID, cfg.Client.Minio.SecretAccessKey, ""),
		Secure: false,
		Region: "",
	})
	if err != nil {
		return err
	}

	if _, err := minioClient.ListBuckets(ctx); err != nil {
		return err
	}

	return nil
}
