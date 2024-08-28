package plugin

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/go-logr/logr"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	apiplugin "github.com/rigdev/rig-go-api/operator/api/v1/plugin"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/obj"
	"github.com/rigdev/rig/pkg/pipeline"
	"github.com/rigdev/rig/pkg/scheme"
	"github.com/rigdev/rig/pkg/uuid"
	"google.golang.org/grpc"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ExecutionContext interface {
	Stop()
	Context() context.Context
}

type executionContext struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func NewExecutionContext(ctx context.Context) ExecutionContext {
	ctx, cancel := context.WithCancel(ctx)
	return &executionContext{
		ctx:    ctx,
		cancel: cancel,
	}
}

func (c *executionContext) Stop() {
	c.cancel()
}

func (c *executionContext) Context() context.Context {
	return c.ctx
}

type pluginExecutor struct {
	context      ExecutionContext
	name         string
	logger       logr.Logger
	client       *plugin.Client
	pluginClient *pluginClient
	binaryPath   string
	args         []string
	tag          string
	id           uuid.UUID
}

func newPluginExecutor(
	context ExecutionContext,
	name, stepTag, pluginTag, pluginConfig, path string,
	args []string,
	logger logr.Logger,
	restConfig *rest.Config,
) (*pluginExecutor, error) {
	tag := stepTag
	if pluginTag != "" {
		tag = pluginTag
	}
	p := &pluginExecutor{
		context:    context,
		name:       name,
		logger:     logger.WithValues("plugin", name),
		binaryPath: path,
		args:       args,
		tag:        tag,
		id:         uuid.New(),
	}

	return p, p.start(context.Context(), pluginConfig, restConfig)
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

func (p *pluginExecutor) start(ctx context.Context, pluginConfig string, restConfig *rest.Config) error {
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
			MagicCookieValue: p.name,
		},
		Plugins: map[string]plugin.Plugin{
			"rigOperatorPlugin": &rigOperatorPlugin{},
		},
		Cmd:              exec.CommandContext(ctx, p.binaryPath, p.args...),
		Logger:           pLogger,
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		SyncStdout:       os.Stderr,
		SyncStderr:       os.Stderr,
		Stderr:           os.Stderr,
	})

	_, err := p.client.Start()
	if err != nil {
		return err
	}

	rpcClient, err := p.client.Client()
	if err != nil {
		return err
	}

	go func() {
		defer p.client.Kill()
		defer p.context.Stop()

		for {
			if p.client.Exited() {
				p.logger.Info("plugin exited")
				return
			}

			if err := rpcClient.Ping(); err != nil {
				p.logger.Error(err, "plugin ping failed")
				return
			}

			time.Sleep(1 * time.Second)
		}
	}()

	raw, err := rpcClient.Dispense("rigOperatorPlugin")
	if err != nil {
		return err
	}

	p.pluginClient = raw.(*pluginClient)

	return p.pluginClient.Initialize(ctx, pluginConfig, p.tag, restConfig)
}

func (p *pluginExecutor) Stop(context.Context) {
	if p.client != nil {
		p.client.Kill()
	}
}

func (p *pluginExecutor) Run(ctx context.Context, req pipeline.CapsuleRequest, opts pipeline.Options) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return p.pluginClient.Run(ctx, req, opts)
}

func (p *pluginExecutor) WatchObjectStatus(
	ctx context.Context,
	namespace string,
	capsule string,
	callback pipeline.ObjectStatusCallback,
) error {
	return p.pluginClient.WatchObjectStatus(ctx, namespace, capsule, callback, p.id)
}

type rigOperatorPlugin struct {
	plugin.NetRPCUnsupportedPlugin
	logger hclog.Logger
	Impl   Plugin
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

func (m *pluginClient) Initialize(ctx context.Context, pluginConfig, tag string, restConfig *rest.Config) error {
	tlsConfigBytes, err := json.Marshal(restConfig.TLSClientConfig)
	if err != nil {
		return err
	}

	restCfg := &apiplugin.RestConfig{
		Host:        restConfig.Host,
		BearerToken: restConfig.BearerToken,
		TlsConfig:   tlsConfigBytes,
	}

	_, err = m.client.Initialize(ctx, &apiplugin.InitializeRequest{
		PluginConfig: pluginConfig,
		Tag:          tag,
		RestConfig:   restCfg,
	})
	return err
}

func (m *pluginClient) Run(ctx context.Context, req pipeline.CapsuleRequest, opts pipeline.Options) error {
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

	var additionalObjects [][]byte
	for _, ao := range opts.AdditionalObjects {
		bs, err := obj.Encode(ao, req.Scheme())
		if err != nil {
			return err
		}
		additionalObjects = append(additionalObjects, bs)
	}

	_, err = m.client.RunCapsule(ctx, &apiplugin.RunCapsuleRequest{
		RunServer:         brokerID,
		CapsuleObject:     capsuleBytes,
		AdditionalObjects: additionalObjects,
	})

	return err
}

func (m *pluginClient) WatchObjectStatus(
	ctx context.Context,
	namespace string,
	capsule string,
	callback pipeline.ObjectStatusCallback,
	pluginID uuid.UUID,
) error {
	c, err := m.client.WatchObjectStatus(ctx, &apiplugin.WatchObjectStatusRequest{
		Namespace: namespace,
		Capsule:   capsule,
	})
	if err != nil {
		return err
	}

	for {
		res, err := c.Recv()
		if err != nil {
			return err
		}

		callback.UpdateStatus(namespace, capsule, pluginID, res.GetChange())
	}
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

	var co client.Object
	var err error
	if req.GetCurrent() {
		if co, err = s.req.GetExisting(gvk, req.GetName()); err != nil {
			return nil, err
		}
	} else {
		if co, err = s.req.GetNew(gvk, req.GetName()); err != nil {
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
	co := obj.New(toGVK(gvk), s.req.Scheme())
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
	if err := s.req.Delete(toGVK(req.GetGvk()), req.GetName()); err != nil {
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

func (s requestServer) ListObjects(
	_ context.Context,
	req *apiplugin.ListObjectsRequest,
) (*apiplugin.ListObjectsResponse, error) {
	var objects []client.Object
	var err error
	if req.GetCurrent() {
		if objects, err = s.req.ListExisting(toGVK(req.GetGvk())); err != nil {
			return nil, err
		}
	} else {
		if objects, err = s.req.ListNew(toGVK(req.GetGvk())); err != nil {
			return nil, err
		}
	}

	var bytes [][]byte
	for _, o := range objects {
		bs, err := obj.Encode(o, s.req.Scheme())
		if err != nil {
			return nil, err
		}
		bytes = append(bytes, bs)

	}

	return &apiplugin.ListObjectsResponse{
		Objects: bytes,
	}, nil
}
