package plugin

import (
	"context"
	"encoding/json"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	apiplugin "github.com/rigdev/rig-go-api/operator/api/v1/plugin"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/obj"
	"github.com/rigdev/rig/pkg/pipeline"
	"github.com/rigdev/rig/pkg/roclient"
	"github.com/rigdev/rig/pkg/scheme"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type GRPCServer struct {
	apiplugin.UnimplementedPluginServiceServer
	logger  hclog.Logger
	Impl    Plugin
	broker  *plugin.GRPCBroker
	scheme  *runtime.Scheme
	watcher Watcher
	cc      client.WithWatch
	vm      scheme.VersionMapper
}

func (m *GRPCServer) Initialize(
	_ context.Context,
	req *apiplugin.InitializeRequest,
) (*apiplugin.InitializeResponse, error) {
	if err := m.Impl.Initialize(InitializeRequest{
		Config: []byte(req.GetPluginConfig()),
		Tag:    req.GetTag(),
		scheme: m.scheme,
	}); err != nil {
		return nil, err
	}

	var restConfig *rest.Config
	var err error
	if req.RestConfig == nil {
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	} else {
		TLSClientConfig := rest.TLSClientConfig{}
		if err = json.Unmarshal(req.RestConfig.TlsConfig, &TLSClientConfig); err != nil {
			return nil, err
		}

		restConfig = &rest.Config{
			Host:            req.RestConfig.Host,
			TLSClientConfig: TLSClientConfig,
			BearerToken:     req.RestConfig.BearerToken,
		}
	}

	cc, err := client.NewWithWatch(restConfig, client.Options{Scheme: m.scheme})
	if err != nil {
		return nil, err
	}

	m.watcher = NewWatcher(m.logger, cc)
	m.cc = cc
	m.vm = scheme.NewVersionMapper(cc)

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
	if capsule.Annotations == nil {
		capsule.Annotations = map[string]string{}
	}
	if capsule.Labels == nil {
		capsule.Labels = map[string]string{}
	}

	reader := roclient.NewReader(m.scheme)
	for _, ao := range req.GetAdditionalObjects() {
		co, err := obj.DecodeAny(ao, m.scheme)
		if err != nil {
			return nil, err
		}

		if err := reader.AddObject(co); err != nil {
			return nil, err
		}
	}

	if err := m.Impl.Run(ctx, &capsuleRequestClient{
		client:  apiplugin.NewRequestServiceClient(conn),
		scheme:  m.scheme,
		capsule: capsule,
		logger:  m.logger,
		ctx:     ctx,
		cc:      m.cc,
		vm:      m.vm,
		cr:      roclient.NewLayeredReader(reader, m.cc),
	}, m.logger); err != nil {
		return nil, err
	}

	return &apiplugin.RunCapsuleResponse{}, nil
}

func (m *GRPCServer) WatchObjectStatus(
	req *apiplugin.WatchObjectStatusRequest,
	stream apiplugin.PluginService_WatchObjectStatusServer,
) error {
	changeChannel := make(chan *apiplugin.ObjectStatusChange, 32)
	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()

	go func() {
		defer cancel()
		for {
			select {
			case <-ctx.Done():
				return
			case change := <-changeChannel:
				if err := stream.Send(&apiplugin.WatchObjectStatusResponse{
					Change: change,
				}); err != nil {
					m.logger.Info("error sending status", "error", err)
					return
				}
			}
		}
	}()

	cw := m.watcher.NewCapsuleWatcher(ctx, req.GetNamespace(), req.GetCapsule(), changeChannel)

	if err := m.Impl.WatchObjectStatus(ctx, cw); err != nil {
		return err
	}

	return nil
}

type capsuleRequestClient struct {
	client  apiplugin.RequestServiceClient
	logger  hclog.Logger
	capsule *v1alpha2.Capsule
	scheme  *runtime.Scheme
	ctx     context.Context
	cc      client.Client
	cr      client.Reader
	vm      scheme.VersionMapper
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

func (c *capsuleRequestClient) Reader() client.Reader {
	return c.cr
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

func fromGK(gk schema.GroupKind) *apiplugin.GVK {
	return &apiplugin.GVK{
		Group: gk.Group,
		Kind:  gk.Kind,
	}
}

func (c *capsuleRequestClient) get(gk schema.GroupKind, name string, current bool) (client.Object, error) {
	res, err := c.client.GetObject(c.ctx, &apiplugin.GetObjectRequest{
		Gvk:     fromGK(gk),
		Name:    name,
		Current: current,
	})
	if err != nil {
		return nil, err
	}

	gvk, err := c.vm.FromGroupKind(gk)
	if err != nil {
		return nil, err
	}

	co := obj.New(gvk, c.scheme)

	if err := obj.DecodeInto(res.GetObject(), co, c.scheme); err != nil {
		return nil, err
	}

	return co, nil
}

func (c *capsuleRequestClient) list(gk schema.GroupKind, current bool) ([]client.Object, error) {
	response, err := c.client.ListObjects(c.ctx, &apiplugin.ListObjectsRequest{
		Gvk:     fromGK(gk),
		Current: current,
	})
	if err != nil {
		return nil, err
	}

	gvk, err := c.vm.FromGroupKind(gk)
	if err != nil {
		return nil, err
	}

	co := obj.New(gvk, c.scheme)
	var res []client.Object
	for _, bytes := range response.GetObjects() {
		obj, err := obj.DecodeIntoT(bytes, co, c.scheme)
		if err != nil {
			return nil, err
		}
		res = append(res, obj)
	}

	return res, nil
}

func (c *capsuleRequestClient) GetExisting(gk schema.GroupKind, name string) (client.Object, error) {
	return c.get(gk, name, true)
}

func (c *capsuleRequestClient) GetExistingInto(obj client.Object) error {
	gvk, err := c.getGVK(obj)
	if err != nil {
		return err
	}

	res, err := c.get(gvk.GroupKind(), obj.GetName(), true)
	if err != nil {
		return err
	}

	return c.scheme.Convert(res, obj, nil)
}

func (c *capsuleRequestClient) GetNew(gk schema.GroupKind, name string) (client.Object, error) {
	return c.get(gk, name, false)
}

func (c *capsuleRequestClient) GetNewInto(obj client.Object) error {
	gvk, err := c.getGVK(obj)
	if err != nil {
		return err
	}

	res, err := c.get(gvk.GroupKind(), obj.GetName(), false)
	if err != nil {
		return err
	}

	return c.scheme.Convert(res, obj, nil)
}

func (c *capsuleRequestClient) ListExisting(gk schema.GroupKind) ([]client.Object, error) {
	return c.list(gk, true)
}

func (c *capsuleRequestClient) ListNew(gk schema.GroupKind) ([]client.Object, error) {
	return c.list(gk, false)
}

func (c *capsuleRequestClient) Set(co client.Object) error {
	gvk, bs, err := c.getGVKAndBytes(co)
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

func (c *capsuleRequestClient) getGVKAndBytes(o client.Object) (schema.GroupVersionKind, []byte, error) {
	gvk, err := c.getGVK(o)
	if err != nil {
		return schema.GroupVersionKind{}, nil, err
	}

	bs, err := obj.Encode(o, c.scheme)
	if err != nil {
		return schema.GroupVersionKind{}, nil, err
	}

	return gvk, bs, nil
}

func (c *capsuleRequestClient) Delete(gk schema.GroupKind, name string) error {
	if _, err := c.client.DeleteObject(c.ctx, &apiplugin.DeleteObjectRequest{
		Gvk:  fromGK(gk),
		Name: name,
	}); err != nil {
		return err
	}

	return nil
}

func (c *capsuleRequestClient) MarkUsedObject(r v1alpha2.UsedResource) error {
	var group string
	if r.Ref.APIGroup != nil {
		group = *r.Ref.APIGroup
	}
	if _, err := c.client.MarkUsedObject(c.ctx, &apiplugin.MarkUsedObjectRequest{
		Gvk: &apiplugin.GVK{
			Group: group,
			Kind:  r.Ref.Kind,
		},
		Name:    r.Ref.Name,
		State:   r.State,
		Message: r.Message,
	}); err != nil {
		return err
	}
	return nil
}

// Plugin is the interface a rig plugin must implement to be used by the rig-operator
type Plugin interface {
	// Run is executed once per reconciliation and throug the CapsuleRequest, has read access
	// to the Capsule being reconciled and read/write access to all other derived resources
	Run(ctx context.Context, req pipeline.CapsuleRequest, logger hclog.Logger) error
	// Initialize is executed once when the rig-operator starts up and is used to pass the configuration
	// of the plugin from the operator to the plugin itself.
	Initialize(req InitializeRequest) error
	WatchObjectStatus(ctx context.Context, watcher CapsuleWatcher) error
}

type NoWatchObjectStatus struct{}

func (NoWatchObjectStatus) WatchObjectStatus(context.Context, CapsuleWatcher) error {
	return errors.UnimplementedErrorf("watch object status not available in plugin")
}

// InitializeRequest contains information needed to initialize the plugin
// This data is constant throughout the execution of the rig-operator.
type InitializeRequest struct {
	Config []byte
	Tag    string

	scheme *runtime.Scheme
}

func (r InitializeRequest) Scheme() *runtime.Scheme {
	return r.scheme
}

// StartPlugin starts the plugin so it can listen for requests to be run on a CapsuleRequest
// name is the name of the plugin as referenced in the rig-operator configuration.
func StartPlugin(name string, rigPlugin Plugin) {
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
