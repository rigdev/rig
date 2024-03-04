package plugin

import (
	"context"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/go-logr/logr"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	apiplugin "github.com/rigdev/rig-go-api/operator/api/v1/plugin"
	"github.com/rigdev/rig/pkg/controller/pipeline"
	"github.com/rigdev/rig/pkg/obj"
	"github.com/rigdev/rig/pkg/scheme"
	"google.golang.org/grpc"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ExternalPlugin struct {
	name         string
	originalName string
	logger       logr.Logger
	client       *plugin.Client
	pluginClient *pluginClient
	path         string
}

func NewExternalPlugin(name, originalName, pluginConfig, path string, logger logr.Logger) (Plugin, error) {
	p := &ExternalPlugin{
		name:         name,
		originalName: originalName,
		logger:       logger,
		path:         path,
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

	p.client = plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: plugin.HandshakeConfig{
			ProtocolVersion:  1,
			MagicCookieKey:   "RIG_OPERATOR_PLUGIN",
			MagicCookieValue: p.originalName,
		},
		Plugins: map[string]plugin.Plugin{
			"rigOperatorPlugin": &rigOperatorPlugin{},
		},
		Cmd:              exec.CommandContext(ctx, p.path),
		Logger:           pLogger,
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
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
) (any, error) {
	return &pluginClient{
		client: apiplugin.NewPluginServiceClient(c),
		broker: broker,
	}, nil
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

	c := make(chan *grpc.Server)
	serverFunc := func(opts []grpc.ServerOption) *grpc.Server {
		s := grpc.NewServer(opts...)
		apiplugin.RegisterRequestServiceServer(s, reqServer)
		c <- s
		return s
	}

	brokerID := m.broker.NextId()
	go m.broker.AcceptAndServe(brokerID, serverFunc)
	s := <-c
	defer s.Stop()

	_, err = m.client.RunCapsule(ctx, &apiplugin.RunCapsuleRequest{
		RunServer:     brokerID,
		CapsuleObject: capsuleBytes,
	})

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
