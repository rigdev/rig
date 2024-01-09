package capabilities_test

import (
	"context"
	"testing"

	"github.com/rigdev/rig-go-api/operator/api/v1/capabilities"
	mockconfig "github.com/rigdev/rig/gen/mocks/github.com/rigdev/rig/pkg/service/config"
	mockdiscovery "github.com/rigdev/rig/gen/mocks/k8s.io/client-go/discovery"
	mockclient "github.com/rigdev/rig/gen/mocks/sigs.k8s.io/controller-runtime/pkg/client"
	"github.com/rigdev/rig/pkg/api/config/v1alpha1"
	svccapabilities "github.com/rigdev/rig/pkg/service/capabilities"
	"github.com/rigdev/rig/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestGet(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		cfg       *v1alpha1.OperatorConfig
		response  *capabilities.GetResponse
		crdErr    error
		apiGroups []metav1.APIGroup
		err       error
	}{
		{
			name:      "if cert manager config is missing ingress is false",
			cfg:       &v1alpha1.OperatorConfig{},
			response:  &capabilities.GetResponse{Ingress: false},
			crdErr:    errors.NewNotFound(schema.GroupResource{}, "oof"),
			apiGroups: []metav1.APIGroup{{Name: "metrics.k8s.io"}, {Name: "some.other.io"}},
			err:       nil,
		},
		{
			name: "if cert manager config is set ingress is true",
			cfg: &v1alpha1.OperatorConfig{
				Certmanager: &v1alpha1.CertManagerConfig{
					ClusterIssuer: "test",
				},
			},
			response: &capabilities.GetResponse{
				Ingress:                     true,
				HasPrometheusServiceMonitor: true,
				HasCustomMetrics:            true,
			},
			crdErr:    nil,
			apiGroups: []metav1.APIGroup{{Name: "metrics.k8s.io"}, {Name: "custom.metrics.k8s.io"}},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mockCfg := mockconfig.NewMockService(t)
			mockClient := mockclient.NewMockClient(t)
			mockDiscovery := mockdiscovery.NewMockDiscoveryInterface(t)

			mockCfg.EXPECT().Operator().Return(tt.cfg)
			mockClient.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(tt.crdErr)
			mockDiscovery.EXPECT().ServerGroups().Return(&metav1.APIGroupList{
				Groups: tt.apiGroups,
			}, nil)

			c := svccapabilities.NewService(mockCfg, mockClient, mockDiscovery)
			res, err := c.Get(context.Background())

			utils.ErrorEqual(t, tt.err, err)
			assert.Equal(t, tt.response, res)
		})
	}
}
