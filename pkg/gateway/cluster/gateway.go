package cluster

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/gen/go/proxy"
	"github.com/rigdev/rig/gen/go/registry"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/iterator"
)

type Capsule struct {
	CapsuleID         string
	Image             string
	ContainerSettings *capsule.ContainerSettings
	Ports             []uint32
	Replicas          uint32
	Volumes           map[string]string
	Network           *capsule.Network
	ConfigFiles       []*capsule.ConfigFile
	Namespace         string
	Metadata          map[string]string
	JWTMethod         *proxy.JWTMethod
	RegistryAuth      *RegistryAuth
}

type RegistryAuth struct {
	Host           string
	RegistrySecret *registry.Secret
}

type Gateway interface {
	ListInstances(ctx context.Context, capsuleID string) (iterator.Iterator[*capsule.Instance], uint64, error)
	RestartInstance(ctx context.Context, capsuleID, instanceID string) error

	Logs(ctx context.Context, capsuleID, instanceID string, follow bool) (iterator.Iterator[*capsule.Log], error)

	ListCapsuleMetrics(ctx context.Context) (iterator.Iterator[*capsule.InstanceMetrics], error)

	CreateVolume(ctx context.Context, id string) error

	// ImageExistsNatively checks if the image exists natively in the cluster. The repo digest is returned if found.
	ImageExistsNatively(ctx context.Context, image string) (bool, string, error)
}

func CreateProxyConfig(ctx context.Context, cn *capsule.Network, jm *proxy.JWTMethod) (*proxy.Config, error) {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return nil, err
	}

	pc := &proxy.Config{
		ProjectId: projectID,
		JwtMethod: jm,
	}

	for _, i := range cn.GetInterfaces() {
		e := &proxy.Interface{
			TargetPort: i.GetPort(),
			Layer:      proxy.Layer_LAYER_4,
		}

		switch v := i.GetPublic().GetMethod().GetKind().(type) {
		case *capsule.RoutingMethod_LoadBalancer_:
			e.SourcePort = v.LoadBalancer.GetPort()
		}

		if i.GetLogging().GetEnabled() {
			e.Layer = proxy.Layer_LAYER_7
			e.Middlewares = append(e.Middlewares, &capsule.Middleware{
				Kind: &capsule.Middleware_Logging{
					Logging: i.GetLogging(),
				},
			})
		}

		if i.GetAuthentication().GetEnabled() {
			e.Layer = proxy.Layer_LAYER_7
			e.Middlewares = append(e.Middlewares, &capsule.Middleware{
				Kind: &capsule.Middleware_Authentication{
					Authentication: i.GetAuthentication(),
				},
			})
		}

		pc.Interfaces = append(pc.Interfaces, e)
	}

	return pc, nil
}
