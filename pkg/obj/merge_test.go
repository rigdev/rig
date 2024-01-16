package obj_test

import (
	"testing"

	"github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/obj"
	"github.com/rigdev/rig/pkg/scheme"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestMerger(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		src      runtime.Object
		dst      runtime.Object
		expected runtime.Object
	}{
		{
			name: "test override",
			src: &v1alpha1.PlatformConfig{
				TypeMeta: v1.TypeMeta{
					Kind:       "PlatformConfig",
					APIVersion: v1alpha1.GroupVersion.String(),
				},
				Auth: v1alpha1.Auth{
					SSO: v1alpha1.SSO{
						OIDCProviders: map[string]v1alpha1.OIDCProvider{
							"test": {
								ClientSecret: "secret",
							},
						},
					},
				},
			},
			dst: &v1alpha1.PlatformConfig{
				TypeMeta: v1.TypeMeta{
					Kind:       "PlatformConfig",
					APIVersion: v1alpha1.GroupVersion.String(),
				},
				Auth: v1alpha1.Auth{
					SSO: v1alpha1.SSO{
						OIDCProviders: map[string]v1alpha1.OIDCProvider{
							"test": {
								ClientID: "id",
							},
						},
					},
				},
			},
			expected: &v1alpha1.PlatformConfig{
				TypeMeta: v1.TypeMeta{
					Kind:       "PlatformConfig",
					APIVersion: v1alpha1.GroupVersion.String(),
				},
				Auth: v1alpha1.Auth{
					SSO: v1alpha1.SSO{
						OIDCProviders: map[string]v1alpha1.OIDCProvider{
							"test": {
								ClientID:     "id",
								ClientSecret: "secret",
							},
						},
					},
				},
			},
		},
	}

	merger := obj.NewMerger(scheme.New())

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := merger.Merge(test.src, test.dst)
			assert.NoError(t, err)

			assert.Equal(t, test.expected, test.dst)
		})
	}
}
