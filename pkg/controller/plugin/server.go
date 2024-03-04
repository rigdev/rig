package plugin

import (
	"context"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	apiplugin "github.com/rigdev/rig-go-api/operator/api/v1/plugin"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/controller/pipeline"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/obj"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type GRPCServer struct {
	apiplugin.UnimplementedPluginServiceServer
	logger hclog.Logger
	Impl   Server
	broker *plugin.GRPCBroker
	scheme *runtime.Scheme
}

func (m *GRPCServer) Initialize(
	_ context.Context,
	req *apiplugin.InitializeRequest,
) (*apiplugin.InitializeResponse, error) {
	if err := m.Impl.LoadConfig([]byte(req.GetPluginConfig())); err != nil {
		return nil, err
	}

	return &apiplugin.InitializeResponse{}, nil
}

func (m *GRPCServer) RunCapsule(
	ctx context.Context,
	req *apiplugin.RunCapsuleRequest,
) (*apiplugin.RunCapsuleResponse, error) {
	conn, err := m.broker.Dial(req.GetRunServer())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	capsule := &v1alpha2.Capsule{}
	if err := obj.DecodeInto(req.CapsuleObject, capsule, m.scheme); err != nil {
		return nil, err
	}

	if err := m.Impl.Run(ctx, &capsuleRequestClient{
		client:  apiplugin.NewRequestServiceClient(conn),
		scheme:  m.scheme,
		capsule: capsule,
		logger:  m.logger,
		ctx:     ctx,
	}, m.logger); err != nil {
		return nil, err
	}

	return &apiplugin.RunCapsuleResponse{}, nil
}

type capsuleRequestClient struct {
	client  apiplugin.RequestServiceClient
	logger  hclog.Logger
	capsule *v1alpha2.Capsule
	scheme  *runtime.Scheme
	ctx     context.Context
}

func (c *capsuleRequestClient) getGVK(obj client.Object) (schema.GroupVersionKind, error) {
	gvks, _, err := c.scheme.ObjectKinds(obj)
	if err != nil {
		c.logger.Error("invalid object type", "error", err)
		return schema.GroupVersionKind{}, err
	}

	return gvks[0], nil
}

func (c *capsuleRequestClient) Scheme() *runtime.Scheme {
	return c.scheme
}

func (c *capsuleRequestClient) Client() client.Client {
	panic("unimplemented `Client` command")
}

func (c *capsuleRequestClient) Capsule() *v1alpha2.Capsule {
	return c.capsule
}

func fromGVK(gvk schema.GroupVersionKind) *apiplugin.GVK {
	return &apiplugin.GVK{
		Group:   gvk.Group,
		Version: gvk.Version,
		Kind:    gvk.Kind,
	}
}

func (c *capsuleRequestClient) get(o client.Object, current bool) error {
	gvk, err := c.getGVK(o)
	if err != nil {
		return err
	}

	res, err := c.client.GetObject(c.ctx, &apiplugin.GetObjectRequest{
		Gvk:     fromGVK(gvk),
		Name:    o.GetName(),
		Current: current,
	})
	if err != nil {
		return err
	}

	return obj.DecodeInto(res.GetObject(), o, c.scheme)
}

func (c *capsuleRequestClient) GetCurrent(obj client.Object) error {
	return c.get(obj, true)
}

func (c *capsuleRequestClient) GetNew(obj client.Object) error {
	return c.get(obj, false)
}

func (c *capsuleRequestClient) Set(co client.Object) error {
	gvk, err := c.getGVK(co)
	if err != nil {
		return err
	}

	bs, err := obj.Encode(co, c.scheme)
	if err != nil {
		return err
	}

	if _, err := c.client.SetObject(c.ctx, &apiplugin.SetObjectRequest{
		Object: bs,
		Gvk:    fromGVK(gvk),
	}); err != nil {
		return err
	}

	return nil
}

func (c *capsuleRequestClient) Delete(_ client.Object) error {
	return errors.UnimplementedErrorf("unimplemented `Delete` command")
}

func (c *capsuleRequestClient) MarkUsedResource(_ v1alpha2.UsedResource) {
	panic("unimplemented `MarkUsedResource` command")
}

type Server interface {
	Run(ctx context.Context, req pipeline.CapsuleRequest, logger hclog.Logger) error
	LoadConfig(data []byte) error
}

func StartPlugin(name string, rigPlugin Server) {
	logger := hclog.New(&hclog.LoggerOptions{
		Level:      hclog.Info,
		Output:     os.Stderr,
		JSONFormat: true,
	}).Named("client")
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: plugin.HandshakeConfig{
			ProtocolVersion:  1,
			MagicCookieKey:   "RIG_OPERATOR_PLUGIN",
			MagicCookieValue: name,
		},
		Plugins: map[string]plugin.Plugin{
			"rigOperatorPlugin": &rigOperatorPlugin{
				Impl:   rigPlugin,
				logger: logger,
			},
		},
		GRPCServer: plugin.DefaultGRPCServer,
		Logger:     logger,
	})
}

func LoadYAMLConfig(data []byte, out any) error {
	return obj.Decode(data, out)
}
