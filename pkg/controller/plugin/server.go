package plugin

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	apiplugin "github.com/rigdev/rig-go-api/operator/api/v1/plugin"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/obj"
	"github.com/rigdev/rig/pkg/pipeline"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type GRPCServer struct {
	apiplugin.UnimplementedPluginServiceServer
	logger hclog.Logger
	Impl   Plugin
	broker *plugin.GRPCBroker
	scheme *runtime.Scheme
}

func (m *GRPCServer) Initialize(
	_ context.Context,
	req *apiplugin.InitializeRequest,
) (*apiplugin.InitializeResponse, error) {
	if err := m.Impl.Initialize(InitializeRequest{
		Config: []byte(req.GetPluginConfig()),
		Tag:    req.GetTag(),
	}); err != nil {
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
	if capsule.Annotations == nil {
		capsule.Annotations = map[string]string{}
	}
	if capsule.Labels == nil {
		capsule.Labels = map[string]string{}
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

func (c *capsuleRequestClient) Reader() client.Reader {
	panic("unimplemented `Reader` command")
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

func (c *capsuleRequestClient) list(o client.Object, current bool) ([]client.Object, error) {
	gvk, err := c.getGVK(o)
	if err != nil {
		return nil, err
	}

	response, err := c.client.ListObjects(c.ctx, &apiplugin.ListObjectsRequest{
		Gvk:     fromGVK(gvk),
		Current: current,
	})
	if err != nil {
		return nil, err
	}

	oo, ok := o.DeepCopyObject().(client.Object)
	if !ok {
		return nil, fmt.Errorf("object had no client.Object type: %T", o)
	}
	var res []client.Object
	for _, bytes := range response.GetObjects() {
		obj, err := obj.DecodeIntoT(bytes, oo, c.scheme)
		if err != nil {
			return nil, err
		}
		res = append(res, obj)
	}

	return res, nil
}

func (c *capsuleRequestClient) GetExisting(obj client.Object) error {
	return c.get(obj, true)
}

func (c *capsuleRequestClient) GetNew(obj client.Object) error {
	return c.get(obj, false)
}

func (c *capsuleRequestClient) ListExisting(obj client.Object) ([]client.Object, error) {
	return c.list(obj, true)
}

func (c *capsuleRequestClient) ListNew(obj client.Object) ([]client.Object, error) {
	return c.list(obj, false)
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

func (c *capsuleRequestClient) Delete(obj client.Object) error {
	gvk, bytes, err := c.getGVKAndBytes(obj)
	if err != nil {
		return err
	}
	if _, err := c.client.DeleteObject(c.ctx, &apiplugin.DeleteObjectRequest{
		Gvk:    fromGVK(gvk),
		Object: bytes,
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
}

// InitializeRequest contains information needed to initialize the plugin
// This data is constant throughout the execution of the rig-operator.
type InitializeRequest struct {
	Config []byte
	Tag    string
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
