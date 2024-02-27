package pipeline

import (
	"context"
	"fmt"

	connect "connectrpc.com/connect"

	apipipeline "github.com/rigdev/rig-go-api/operator/api/v1/pipeline"
	"github.com/rigdev/rig-go-api/operator/api/v1/pipeline/pipelineconnect"
	"github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/controller/pipeline"
	"github.com/rigdev/rig/pkg/obj"
	"github.com/rigdev/rig/pkg/scheme"
	svcpipeline "github.com/rigdev/rig/pkg/service/pipeline"
)

func NewHandler(pipeline svcpipeline.Service) pipelineconnect.ServiceHandler {
	return &handler{pipeline: pipeline}
}

type handler struct {
	pipeline svcpipeline.Service
}

func (h *handler) DryRun(
	ctx context.Context,
	req *connect.Request[apipipeline.DryRunRequest],
) (*connect.Response[apipipeline.DryRunResponse], error) {
	scheme := scheme.New()

	var cfg *v1alpha1.OperatorConfig
	if req.Msg.GetOperatorConfig() != "" {
		cfg = &v1alpha1.OperatorConfig{}
		if err := obj.DecodeInto([]byte(req.Msg.GetOperatorConfig()), cfg, scheme); err != nil {
			return nil, err
		}

		cfg.Default()
	}

	var spec *v1alpha2.Capsule
	if req.Msg.GetCapsuleSpec() != "" {
		spec = &v1alpha2.Capsule{}
		if err := obj.DecodeInto([]byte(req.Msg.GetCapsuleSpec()), spec, scheme); err != nil {
			return nil, err
		}
	}

	var opts []pipeline.CapsuleRequestOption
	if req.Msg.GetForce() {
		opts = append(opts, pipeline.WithForce())
	}

	result, err := h.pipeline.DryRun(ctx, cfg, req.Msg.GetNamespace(), req.Msg.GetCapsule(), spec, opts...)
	if err != nil {
		fmt.Println("failed to dry run it", err)
		return nil, err
	}

	res := &apipipeline.DryRunResponse{}
	for _, o := range result.InputObjects {
		bs, err := obj.Encode(o, scheme)
		if err != nil {
			return nil, err
		}

		res.InputObjects = append(res.InputObjects, &apipipeline.Object{
			Gvk: &apipipeline.GVK{
				Group:   o.GetObjectKind().GroupVersionKind().Group,
				Version: o.GetObjectKind().GroupVersionKind().Version,
				Kind:    o.GetObjectKind().GroupVersionKind().Kind,
			},
			Name:    o.GetName(),
			Content: string(bs),
		})
	}

	for _, oo := range result.OutputObjects {
		o := &apipipeline.Object{
			Gvk: &apipipeline.GVK{
				Group:   oo.ObjectKey.Group,
				Version: oo.ObjectKey.Version,
				Kind:    oo.ObjectKey.Kind,
			},
			Name: oo.ObjectKey.Name,
		}
		if oo.Object != nil {
			bs, err := obj.Encode(oo.Object, scheme)
			if err != nil {
				return nil, err
			}

			o.Content = string(bs)
		}

		var state apipipeline.ObjectState
		switch oo.State {
		case pipeline.ResourceStateAlreadyExists:
			state = apipipeline.ObjectState_OBJECT_STATE_ALREADY_EXISTS
		case pipeline.ResourceStateCreated:
			state = apipipeline.ObjectState_OBJECT_STATE_CREATE
		case pipeline.ResourceStateDeleted:
			state = apipipeline.ObjectState_OBJECT_STATE_DELETE
		case pipeline.ResourceStateUpdated:
			state = apipipeline.ObjectState_OBJECT_STATE_UPDATE
		case pipeline.ResourceStateUnchanged:
			state = apipipeline.ObjectState_OBJECT_STATE_UNCHANGED
		}

		res.OutputObjects = append(res.OutputObjects, &apipipeline.ObjectChange{
			Object: o,
			State:  state,
		})
	}

	return connect.NewResponse(res), nil
}
