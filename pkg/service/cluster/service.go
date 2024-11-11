package cluster

import (
	"context"
	"fmt"
	"maps"
	"slices"

	"github.com/rigdev/rig-go-api/model"
	api_cluster "github.com/rigdev/rig-go-api/operator/api/v1/cluster"
	"github.com/rigdev/rig/pkg/pipeline"
	corev1 "k8s.io/api/core/v1"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Service interface {
	GetNodes(ctx context.Context) ([]*api_cluster.Node, error)
	GetNodePods(ctx context.Context, nodeName string) ([]*api_cluster.Pod, error)
}

func New(client client.Client) Service {
	return &service{
		client: client,
	}
}

type service struct {
	client client.Client
}

func (s *service) GetNodes(ctx context.Context) ([]*api_cluster.Node, error) {
	listReq := corev1.NodeList{}
	if err := s.client.List(ctx, &listReq); err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	nodes := map[string]*api_cluster.Node{}

	for _, node := range listReq.Items {
		nodes[node.GetName()] = &api_cluster.Node{
			NodeName: node.GetName(),
			Allocateable: &model.Resources{
				CpuMillis:   uint64(node.Status.Allocatable.Cpu().MilliValue()),
				MemoryBytes: uint64(node.Status.Allocatable.Memory().Value()),
				Pods:        uint64(node.Status.Allocatable.Pods().Value()),
			},
		}
	}

	list := metricsv1beta1.NodeMetricsList{}
	if err := s.client.List(ctx, &list); err != nil {
		return nil, fmt.Errorf("failed to list node metrics: %w", err)
	}

	for _, node := range list.Items {
		n, ok := nodes[node.GetName()]
		if !ok {
			n = &api_cluster.Node{
				NodeName:     node.GetName(),
				Allocateable: &model.Resources{},
				Usage:        &model.Resources{},
			}
			nodes[node.GetName()] = n
		}
		n.Usage = &model.Resources{
			CpuMillis:   uint64(node.Usage.Cpu().MilliValue()),
			MemoryBytes: uint64(node.Usage.Memory().Value()),
			Pods:        uint64(node.Usage.Pods().Value()),
		}
	}

	keys := slices.Sorted((maps.Keys(nodes)))
	var res []*api_cluster.Node
	for _, k := range keys {
		res = append(res, nodes[k])
	}
	return res, nil
}

func (s *service) GetNodePods(ctx context.Context, nodeName string) ([]*api_cluster.Pod, error) {
	listReq := corev1.PodList{}
	if err := s.client.List(ctx, &listReq, client.MatchingFields{
		"spec.nodeName": nodeName,
	}); err != nil {
		return nil, err
	}

	var res []*api_cluster.Pod
	for _, pod := range listReq.Items {
		req := &model.Resources{}
		for _, c := range pod.Spec.Containers {
			req.CpuMillis += uint64(c.Resources.Requests.Cpu().MilliValue())
			req.MemoryBytes += uint64(c.Resources.Requests.Memory().Value())
		}
		res = append(res, &api_cluster.Pod{
			PodName:     pod.GetName(),
			Namespace:   pod.GetNamespace(),
			Requested:   req,
			CapsuleName: pod.Labels[pipeline.LabelCapsule],
		})
	}

	return res, nil
}
