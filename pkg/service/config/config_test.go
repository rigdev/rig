package config

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/scheme"
	"github.com/rigdev/rig/pkg/utils"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewPlatformConfig(t *testing.T) {
	tests := []struct {
		name        string
		files       []string
		envVars     map[string]string
		expected    *v1alpha1.PlatformConfig
		expectedErr error
	}{
		{
			name: "one file",
			files: []string{
				`apiversion: config.rig.dev/v1alpha1
kind: PlatformConfig
port: 1234
publicURL: hej.com`,
			},
			expected: &v1alpha1.PlatformConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       "PlatformConfig",
					APIVersion: "config.rig.dev/v1alpha1",
				},
				TelemetryEnabled: true,
				Port:             1234,
				PublicURL:        "hej.com",
				Client: v1alpha1.Client{
					Mailjets: map[string]v1alpha1.ClientMailjet{},
					SMTPs:    map[string]v1alpha1.ClientSMTP{},
					Postgres: v1alpha1.ClientPostgres{
						Port:     5432,
						Database: "rig",
					},
					Operator: v1alpha1.ClientOperator{
						BaseURL: "rig-operator:9000",
					},
				},
				Repository: v1alpha1.Repository{
					Store: "postgres",
				},
				DockerRegistries: map[string]v1alpha1.DockerRegistryCredentials{},
			},
		},
		{
			name: "multiple files",
			files: []string{
				`apiversion: config.rig.dev/v1alpha1
kind: PlatformConfig
port: 1234
publicURL: hej.com`,
				`apiversion: config.rig.dev/v1alpha1
kind: PlatformConfig
port: 1235
auth:
  secret: "123"
`,
			},
			expected: &v1alpha1.PlatformConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       "PlatformConfig",
					APIVersion: "config.rig.dev/v1alpha1",
				},
				Port:             1235,
				PublicURL:        "hej.com",
				Auth:             v1alpha1.Auth{Secret: "123"},
				TelemetryEnabled: true,
				Client: v1alpha1.Client{
					Mailjets: map[string]v1alpha1.ClientMailjet{},
					SMTPs:    map[string]v1alpha1.ClientSMTP{},
					Postgres: v1alpha1.ClientPostgres{
						Port:     5432,
						Database: "rig",
					},
					Operator: v1alpha1.ClientOperator{
						BaseURL: "rig-operator:9000",
					},
				},
				Repository: v1alpha1.Repository{
					Store: "postgres",
				},
				DockerRegistries: map[string]v1alpha1.DockerRegistryCredentials{},
			},
		},
		{
			name: "env vars",
			files: []string{
				`apiversion: config.rig.dev/v1alpha1
kind: PlatformConfig
port: 1234
publicURL: hej.com`,
			},
			envVars: map[string]string{
				"RIG_AUTH_SECRET": "secret",
				"RIG_PUBLICURL":   "hej2.com",
			},
			expected: &v1alpha1.PlatformConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       "PlatformConfig",
					APIVersion: "config.rig.dev/v1alpha1",
				},
				Port:             1234,
				PublicURL:        "hej2.com",
				Auth:             v1alpha1.Auth{Secret: "secret"},
				TelemetryEnabled: true,
				Client: v1alpha1.Client{
					Mailjets: map[string]v1alpha1.ClientMailjet{},
					SMTPs:    map[string]v1alpha1.ClientSMTP{},
					Postgres: v1alpha1.ClientPostgres{
						Port:     5432,
						Database: "rig",
					},
					Operator: v1alpha1.ClientOperator{
						BaseURL: "rig-operator:9000",
					},
				},
				Repository: v1alpha1.Repository{
					Store: "postgres",
				},
				DockerRegistries: map[string]v1alpha1.DockerRegistryCredentials{},
			},
		},
		{
			name: "error: bad validation",
			files: []string{
				`apiVersion: config.rig.dev/v1alpha1
kind: PlatformConfig
port: 1234
publicURL: hej.com
capsuleExtensions:
  key:
    schema:
      type: string`,
			},
			expected: nil,
			expectedErr: errors.New(
				"capsuleExtensiosn[key].schema.type: Invalid value: \"string\": top level schema must be of type 'object'",
			),
		},
	}

	scheme := scheme.New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			var filePaths []string
			for idx, file := range tt.files {
				path := fmt.Sprintf("/file%v.yaml", idx)
				err := afero.WriteFile(fs, path, []byte(file), os.ModePerm)
				require.NoError(t, err)
				filePaths = append(filePaths, path)
			}

			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			cfg, err := NewPlatformConfig(
				fs, scheme, WithFilePaths(filePaths...),
			)
			utils.ErrorEqual(t, tt.expectedErr, err)
			require.Equal(t, tt.expected, cfg)
		})
	}
}
