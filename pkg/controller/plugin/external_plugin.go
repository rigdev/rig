package plugin

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/go-logr/logr"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	apiplugin "github.com/rigdev/rig-go-api/operator/api/v1/plugin"
	"github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/controller/pipeline"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/obj"
	"github.com/rigdev/rig/pkg/scheme"
	"google.golang.org/grpc"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ExternalPlugin struct {
	name         string
	logger       logr.Logger
	client       *plugin.Client
	pluginClient *pluginClient
}

func NewExternalPlugin(name string, logger logr.Logger, pluginConfig string) (Plugin, error) {
	p := &ExternalPlugin{
		name:   name,
		logger: logger,
	}

	return p, p.start(context.Background(), pluginConfig)
}

type loggerSink struct {
	logger logr.Logger
}

func (l *loggerSink) Accept(name string, level hclog.Level, msg string, args ...interface{}) {
	logger := l.logger.WithName(name).WithValues(args...)
	if level < hclog.Info {
		return
	}
	logger.Info(msg)
}

func (p *ExternalPlugin) start(ctx context.Context, pluginConfig string) error {
	pLogger := hclog.NewInterceptLogger(&hclog.LoggerOptions{
		Name:       p.name,
		Output:     io.Discard,
		Level:      hclog.Info,
		JSONFormat: true,
	})
	pLogger.RegisterSink(&loggerSink{
		logger: p.logger.WithName("plugin"),
	})

	execPath, err := os.Executable()
	if err != nil {
		return err
	}

	pluginDir := path.Join(path.Dir(execPath), "plugin")
	if dir := os.Getenv("RIG_PLUGIN_DIR"); dir != "" {
		pluginDir = dir
	}

	p.client = plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: plugin.HandshakeConfig{
			ProtocolVersion:  1,
			MagicCookieKey:   "RIG_OPERATOR_PLUGIN",
			MagicCookieValue: p.name,
		},
		Plugins: map[string]plugin.Plugin{
			"rigOperatorPlugin": &rigOperatorPlugin{},
		},
		Cmd:              exec.CommandContext(ctx, path.Join(pluginDir, p.name)),
		Logger:           pLogger,
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		SyncStdout:       os.Stdout,
		SyncStderr:       os.Stderr,
		Stderr:           os.Stderr,
	})

	rpcClient, err := p.client.Client()
	if err != nil {
		return err
	}

	raw, err := rpcClient.Dispense("rigOperatorPlugin")
	if err != nil {
		return err
	}

	p.pluginClient = raw.(*pluginClient)

	return p.pluginClient.Initialize(ctx, pluginConfig)
}

func (p *ExternalPlugin) Stop(context.Context) {
	if p.client != nil {
		p.client.Kill()
	}
}

func (p *ExternalPlugin) Run(ctx context.Context, req pipeline.CapsuleRequest) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return p.pluginClient.Run(ctx, req)
}

type rigOperatorPlugin struct {
	plugin.NetRPCUnsupportedPlugin
	logger hclog.Logger
	Impl   Server
}

func (p *rigOperatorPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	apiplugin.RegisterPluginServiceServer(s, &GRPCServer{
		Impl:   p.Impl,
		logger: p.logger,
		broker: broker,
		scheme: scheme.New(),
	})
	return nil
}

func (p *rigOperatorPlugin) GRPCClient(
	_ context.Context,
	broker *plugin.GRPCBroker,
	c *grpc.ClientConn,
) (interface{}, error) {
	return &pluginClient{
		client: apiplugin.NewPluginServiceClient(c),
		broker: broker,
	}, nil
}

type GRPCServer struct {
	apiplugin.UnimplementedPluginServiceServer
	logger         hclog.Logger
	Impl           Server
	broker         *plugin.GRPCBroker
	operatorConfig v1alpha1.OperatorConfig
	scheme         *runtime.Scheme
}

func (m *GRPCServer) Initialize(
	_ context.Context,
	req *apiplugin.InitializeRequest,
) (*apiplugin.InitializeResponse, error) {
	if err := m.Impl.LoadConfig([]byte(req.GetPluginConfig())); err != nil {
		return nil, err
	}

	// if err := obj.DecodeInto(req.OperatorConfig, &m.operatorConfig, m.scheme); err != nil {
	// 	return nil, err
	// }
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
		client:         apiplugin.NewRequestServiceClient(conn),
		scheme:         m.scheme,
		operatorConfig: &m.operatorConfig,
		capsule:        capsule,
		logger:         m.logger,
		ctx:            ctx,
	}, m.logger); err != nil {
		return nil, err
	}

	return &apiplugin.RunCapsuleResponse{}, nil
}

type pluginClient struct {
	broker *plugin.GRPCBroker
	client apiplugin.PluginServiceClient
}

func (m *pluginClient) Initialize(ctx context.Context, pluginConfig string) error {
	_, err := m.client.Initialize(ctx, &apiplugin.InitializeRequest{
		PluginConfig: pluginConfig,
	})
	return err
}

func (m *pluginClient) Run(ctx context.Context, req pipeline.CapsuleRequest) error {
	reqServer := &requestServer{req: req}

	capsuleBytes, err := obj.Encode(req.Capsule(), req.Scheme())
	if err != nil {
		return err
	}

	var s *grpc.Server
	serverFunc := func(opts []grpc.ServerOption) *grpc.Server {
		s = grpc.NewServer(opts...)
		apiplugin.RegisterRequestServiceServer(s, reqServer)
		return s
	}

	brokerID := m.broker.NextId()
	go m.broker.AcceptAndServe(brokerID, serverFunc)

	_, err = m.client.RunCapsule(ctx, &apiplugin.RunCapsuleRequest{
		RunServer:     brokerID,
		CapsuleObject: capsuleBytes,
	})

	s.Stop()

	return err
}

type requestServer struct {
	apiplugin.UnimplementedRequestServiceServer

	req pipeline.CapsuleRequest
}

func toGVK(gvk *apiplugin.GVK) schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   gvk.GetGroup(),
		Version: gvk.GetVersion(),
		Kind:    gvk.GetKind(),
	}
}

func (s requestServer) GetObject(
	_ context.Context,
	req *apiplugin.GetObjectRequest,
) (*apiplugin.GetObjectResponse, error) {
	gvk := toGVK(req.GetGvk())
	ro, err := s.req.Scheme().New(gvk)
	if err != nil {
		return nil, err
	}

	co := ro.(client.Object)
	co.SetName(req.GetName())
	if req.GetCurrent() {
		if err := s.req.GetCurrent(co); err != nil {
			return nil, err
		}
	} else {
		if err := s.req.GetNew(co); err != nil {
			return nil, err
		}
	}

	bs, err := obj.Encode(co, s.req.Scheme())
	if err != nil {
		return nil, err
	}

	return &apiplugin.GetObjectResponse{
		Object: bs,
	}, nil
}

func (s requestServer) SetObject(
	_ context.Context,
	req *apiplugin.SetObjectRequest,
) (*apiplugin.SetObjectResponse, error) {
	gvk := toGVK(req.GetGvk())
	ro, err := s.req.Scheme().New(gvk)
	if err != nil {
		return nil, err
	}

	co := ro.(client.Object)
	if err := obj.DecodeInto(req.GetObject(), co, s.req.Scheme()); err != nil {
		return nil, err
	}

	if err := s.req.Set(co); err != nil {
		return nil, err
	}

	return &apiplugin.SetObjectResponse{}, nil
}

type capsuleRequestClient struct {
	client         apiplugin.RequestServiceClient
	logger         hclog.Logger
	operatorConfig *v1alpha1.OperatorConfig
	capsule        *v1alpha2.Capsule
	scheme         *runtime.Scheme
	ctx            context.Context
}

func (c *capsuleRequestClient) getGVK(obj client.Object) (schema.GroupVersionKind, error) {
	gvks, _, err := c.scheme.ObjectKinds(obj)
	if err != nil {
		c.logger.Error("invalid object type", "error", err)
		return schema.GroupVersionKind{}, err
	}

	return gvks[0], nil
}

func (c *capsuleRequestClient) Config() *v1alpha1.OperatorConfig {
	return c.operatorConfig
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

func LoadYAMLConfig(data []byte, out interface{}) error {
	return obj.DecodeYAML(data, out)
}
