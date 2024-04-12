package mod

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
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/controller/pipeline"
	"github.com/rigdev/rig/pkg/obj"
	"github.com/rigdev/rig/pkg/scheme"
	"google.golang.org/grpc"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type modExecutor struct {
	name       string
	logger     logr.Logger
	client     *plugin.Client
	modClient  *modClient
	binaryPath string
	args       []string
	tag        string
}

func newModExecutor(
	name, stepTag, modTag, modConfig, path string,
	args []string,
	logger logr.Logger,
) (*modExecutor, error) {
	tag := stepTag
	if modTag != "" {
		tag = modTag
	}
	p := &modExecutor{
		name:       name,
		logger:     logger,
		binaryPath: path,
		args:       args,
		tag:        tag,
	}

	return p, p.start(context.Background(), modConfig)
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

func (e *modExecutor) start(ctx context.Context, modConfig string) error {
	pLogger := hclog.NewInterceptLogger(&hclog.LoggerOptions{
		Name:       e.name,
		Output:     io.Discard,
		Level:      hclog.Info,
		JSONFormat: true,
	})
	pLogger.RegisterSink(&loggerSink{
		logger: e.logger.WithName("mod"),
	})

	e.client = plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: plugin.HandshakeConfig{
			ProtocolVersion:  1,
			MagicCookieKey:   "RIG_OPERATOR_MOD",
			MagicCookieValue: e.name,
		},
		Plugins: map[string]plugin.Plugin{
			"rigOperatorMod": &rigOperatorMod{},
		},
		Cmd:              exec.CommandContext(ctx, e.binaryPath, e.args...),
		Logger:           pLogger,
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		Stderr:           os.Stderr,
	})

	rpcClient, err := e.client.Client()
	if err != nil {
		return err
	}

	raw, err := rpcClient.Dispense("rigOperatorMod")
	if err != nil {
		return err
	}

	e.modClient = raw.(*modClient)

	return e.modClient.Initialize(ctx, modConfig, e.tag)
}

func (e *modExecutor) Stop(context.Context) {
	if e.client != nil {
		e.client.Kill()
	}
}

func (e *modExecutor) Run(ctx context.Context, req pipeline.CapsuleRequest) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return e.modClient.Run(ctx, req)
}

type rigOperatorMod struct {
	plugin.NetRPCUnsupportedPlugin
	logger hclog.Logger
	Impl   Mod
}

func (m *rigOperatorMod) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	apiplugin.RegisterPluginServiceServer(s, &GRPCServer{
		Impl:   m.Impl,
		logger: m.logger,
		broker: broker,
		scheme: scheme.New(),
	})
	return nil
}

func (m *rigOperatorMod) GRPCClient(
	_ context.Context,
	broker *plugin.GRPCBroker,
	c *grpc.ClientConn,
) (any, error) {
	return &modClient{
		client: apiplugin.NewPluginServiceClient(c),
		broker: broker,
	}, nil
}

type modClient struct {
	broker *plugin.GRPCBroker
	client apiplugin.PluginServiceClient
}

func (c *modClient) Initialize(ctx context.Context, modConfig, tag string) error {
	_, err := c.client.Initialize(ctx, &apiplugin.InitializeRequest{
		PluginConfig: modConfig,
		Tag:          tag,
	})
	return err
}

func (c *modClient) Run(ctx context.Context, req pipeline.CapsuleRequest) error {
	reqServer := &requestServer{req: req}

	capsuleBytes, err := obj.Encode(req.Capsule(), req.Scheme())
	if err != nil {
		return err
	}

	sc := make(chan *grpc.Server)
	serverFunc := func(opts []grpc.ServerOption) *grpc.Server {
		s := grpc.NewServer(opts...)
		apiplugin.RegisterRequestServiceServer(s, reqServer)
		sc <- s
		return s
	}

	brokerID := c.broker.NextId()
	go c.broker.AcceptAndServe(brokerID, serverFunc)
	s := <-sc
	defer s.Stop()

	_, err = c.client.RunCapsule(ctx, &apiplugin.RunCapsuleRequest{
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
		if err := s.req.GetExisting(co); err != nil {
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

func (s requestServer) decodeObject(gvk *apiplugin.GVK, bytes []byte) (client.Object, error) {
	ro, err := s.req.Scheme().New(toGVK(gvk))
	if err != nil {
		return nil, err
	}

	co := ro.(client.Object)
	if err := obj.DecodeInto(bytes, co, s.req.Scheme()); err != nil {
		return nil, err
	}

	return co, nil
}

func (s requestServer) SetObject(
	_ context.Context,
	req *apiplugin.SetObjectRequest,
) (*apiplugin.SetObjectResponse, error) {
	obj, err := s.decodeObject(req.GetGvk(), req.GetObject())
	if err != nil {
		return nil, err
	}
	if err := s.req.Set(obj); err != nil {
		return nil, err
	}

	return &apiplugin.SetObjectResponse{}, nil
}

func (s requestServer) Delete(
	_ context.Context,
	req *apiplugin.DeleteObjectRequest,
) (*apiplugin.DeleteObjectResponse, error) {
	obj, err := s.decodeObject(req.GetGvk(), req.GetObject())
	if err != nil {
		return nil, err
	}
	if err := s.req.Delete(obj); err != nil {
		return nil, err
	}

	return &apiplugin.DeleteObjectResponse{}, nil
}

func (s requestServer) MarkUsedObject(
	_ context.Context,
	req *apiplugin.MarkUsedObjectRequest,
) (*apiplugin.MarkUsedObjectResponse, error) {
	var group *string
	if g := req.GetGvk().GetGroup(); g != "" {
		group = &g
	}
	if err := s.req.MarkUsedObject(v1alpha2.UsedResource{
		Ref: &v1.TypedLocalObjectReference{
			APIGroup: group,
			Kind:     req.GetGvk().GetKind(),
			Name:     req.GetName(),
		},
		State:   req.GetState(),
		Message: req.GetMessage(),
	}); err != nil {
		return nil, err
	}
	return &apiplugin.MarkUsedObjectResponse{}, nil
}
