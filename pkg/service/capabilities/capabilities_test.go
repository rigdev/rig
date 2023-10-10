package capabilities_test

import (
	"context"
	"testing"

	"github.com/rigdev/rig-go-api/operator/api/v1/capabilities"
	"github.com/rigdev/rig/pkg/api/config/v1alpha1"
	svccapabilities "github.com/rigdev/rig/pkg/service/capabilities"
	"github.com/stretchr/testify/assert"
)

func newMockConfig() *mockConfig {
	cfg := &v1alpha1.OperatorConfig{}
	cfg.Default()
	return &mockConfig{cfg: cfg}
}

type mockConfig struct {
	cfg *v1alpha1.OperatorConfig
}

func (c *mockConfig) Get() *v1alpha1.OperatorConfig {
	return c.cfg
}

func TestGet(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		cfg      *v1alpha1.OperatorConfig
		response *capabilities.GetResponse
		err      error
	}{
		{
			name: "if cert manager config is missing ingress is false",
			response: &capabilities.GetResponse{
				Ingress: false,
			},
		},
		{
			name: "if cert manager config is set ingress is true",
			cfg: &v1alpha1.OperatorConfig{
				Certmanager: &v1alpha1.CertManagerConfig{
					ClusterIssuer: "test",
				},
			},
			response: &capabilities.GetResponse{
				Ingress: true,
			},
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			cfg := newMockConfig()
			if test.cfg != nil {
				cfg.cfg = test.cfg
			}
			c := svccapabilities.NewService(cfg)

			res, err := c.Get(context.Background(), &capabilities.GetRequest{})

			if test.err == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorAs(t, err, &test.err)
			}
			assert.Equal(t, test.response, res)
		})
	}
}
