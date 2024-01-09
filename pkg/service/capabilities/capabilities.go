package capabilities

import (
	"context"

	"github.com/rigdev/rig-go-api/operator/api/v1/capabilities"
	"github.com/rigdev/rig/pkg/service/config"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Service interface {
	Get(ctx context.Context) (*capabilities.GetResponse, error)
}

func NewService(
	cfg config.Service,
	client client.Client,
	discoveryClient discovery.DiscoveryInterface,
) Service {
	return &service{
		cfg:             cfg,
		client:          client,
		discoveryClient: discoveryClient,
	}
}

type service struct {
	cfg             config.Service
	client          client.Client
	discoveryClient discovery.DiscoveryInterface
}

// Get implements Service.
func (s *service) Get(ctx context.Context) (*capabilities.GetResponse, error) {
	res := &capabilities.GetResponse{}

	cfg := s.cfg.Operator()
	if cfg.Certmanager != nil && cfg.Certmanager.ClusterIssuer != "" {
		res.Ingress = true
	}

	ok, err := s.hasServiceMonitor(ctx)
	if err != nil {
		return nil, err
	}
	res.HasPrometheusServiceMonitor = ok

	ok, err = s.hasCustomMetricsAPI()
	if err != nil {
		return nil, err
	}
	res.HasCustomMetrics = ok

	return res, nil
}

func (s *service) hasServiceMonitor(ctx context.Context) (bool, error) {
	crdList := &apiextensionsv1.CustomResourceDefinitionList{}
	if err := s.client.List(ctx, crdList); err != nil {
		return false, err
	}
	for _, crd := range crdList.Items {
		if crd.Name == "servicemonitors.monitoring.coreos.com" {
			return true, nil
		}
	}
	return false, nil
}

func (s *service) hasCustomMetricsAPI() (bool, error) {
	groups, err := s.discoveryClient.ServerGroups()
	if err != nil {
		return false, err
	}

	for _, g := range groups.Groups {
		if g.Name == "custom.metrics.k8s.io" {
			return true, nil
		}
	}

	return false, nil
}
