package pipeline

import (
	"context"

	connect "connectrpc.com/connect"
	apipipeline "github.com/rigdev/rig-go-api/operator/api/v1/pipeline"
	"github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/obj"
	"github.com/rigdev/rig/pkg/pipeline"
)

func (h *handler) decodeOperatorConfig(config string) (*v1alpha1.OperatorConfig, error) {
	if config == "" {
		return nil, nil
	}
	cfg := &v1alpha1.OperatorConfig{}
	cfg, err := obj.DecodeIntoT([]byte(config), cfg, h.scheme)
	if err != nil {
		return nil, err
	}
	cfg.Default()

	return cfg, nil
}

func (h *handler) DryRun(
	ctx context.Context,
	req *connect.Request[apipipeline.DryRunRequest],
) (*connect.Response[apipipeline.DryRunResponse], error) {
	cfg, err := h.decodeOperatorConfig(req.Msg.GetOperatorConfig())
	if err != nil {
		return nil, err
	}

	var spec *v1alpha2.Capsule
	if req.Msg.GetCapsuleSpec() != "" {
		spec = &v1alpha2.Capsule{}
		if err := obj.DecodeInto([]byte(req.Msg.GetCapsuleSpec()), spec, h.scheme); err != nil {
			return nil, err
		}
	}

	var opts []pipeline.CapsuleRequestOption
	if req.Msg.GetForce() {
		opts = append(opts, pipeline.WithForce())
	}

	if len(req.Msg.GetAdditionalObjects()) > 0 {
		opts = append(opts, pipeline.WithAdditionalResources(req.Msg.AdditionalObjects))
	}

	result, err := h.pipeline.DryRun(ctx, cfg, req.Msg.GetNamespace(), req.Msg.GetCapsule(), spec, opts...)
	if err != nil {
		return nil, err
	}

	res := &apipipeline.DryRunResponse{}
	for _, o := range result.InputObjects {
		bs, err := obj.Encode(o, h.scheme)
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
			bs, err := obj.Encode(oo.Object, h.scheme)
			if err != nil {
				return nil, err
			}

			o.Content = string(bs)
		}

		var outcome apipipeline.ObjectOutcome
		switch oo.State {
		case pipeline.ResourceStateAlreadyExists:
			outcome = apipipeline.ObjectOutcome_OBJECT_OUTCOME_ALREADY_EXISTS
		case pipeline.ResourceStateCreated:
			outcome = apipipeline.ObjectOutcome_OBJECT_OUTCOME_CREATE
		case pipeline.ResourceStateDeleted:
			outcome = apipipeline.ObjectOutcome_OBJECT_OUTCOME_DELETE
		case pipeline.ResourceStateUpdated:
			outcome = apipipeline.ObjectOutcome_OBJECT_OUTCOME_UPDATE
		case pipeline.ResourceStateUnchanged:
			outcome = apipipeline.ObjectOutcome_OBJECT_OUTCOME_UNCHANGED
		}

		res.OutputObjects = append(res.OutputObjects, &apipipeline.ObjectChange{
			Object:  o,
			Outcome: outcome,
		})
	}

	return connect.NewResponse(res), nil
}
