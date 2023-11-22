package platform

import (
	"fmt"
	"os"
	"testing"

	"github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/manager"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name             string
		fileConfig       []string
		secretFileConfig []string
		envVars          map[string]string
		err              error
		expected         func() *v1alpha1.PlatformConfig
	}{
		{
			name:     "nothing set returns default",
			expected: v1alpha1.NewDefaultPlatform,
		},
		{
			name: "env overrides default",
			envVars: map[string]string{
				"RIG_PORT": "4242",
			},
			expected: func() *v1alpha1.PlatformConfig {
				c := v1alpha1.NewDefaultPlatform()
				c.Port = 4242
				return c
			},
		},
		{
			name: "config is read from file",
			fileConfig: []string{
				`port: 4242`,
			},
			expected: func() *v1alpha1.PlatformConfig {
				c := v1alpha1.NewDefaultPlatform()
				c.Port = 4242
				return c
			},
		},
		{
			name: "secrets is read from file",
			secretFileConfig: []string{
				`repository: 
    secret: test`,
			},
			expected: func() *v1alpha1.PlatformConfig {
				c := v1alpha1.NewDefaultPlatform()
				c.Repository.Secret = "test"
				return c
			},
		},
		{
			name: "secrets overwrites config",
			fileConfig: []string{
				`repository:
    secret: test`,
			},
			secretFileConfig: []string{
				`repository:
    secret: test2`,
			},
			expected: func() *v1alpha1.PlatformConfig {
				c := v1alpha1.NewDefaultPlatform()
				c.Repository.Secret = "test2"
				return c
			},
		},
		{
			name: "env is not read if secret is specified",
			envVars: map[string]string{
				"RIG_PORT": "4242",
			},
			secretFileConfig: []string{
				`port: 4243`,
			},
			expected: func() *v1alpha1.PlatformConfig {
				c := v1alpha1.NewDefaultPlatform()
				c.Port = 4243
				return c
			},
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			var f *os.File
			var fs *os.File
			var err error
			if len(test.fileConfig) > 0 {
				f, err = os.CreateTemp("", "rig_test*.yaml")
				assert.NoError(t, err)
				_, err := f.WriteString("apiVersion: config.rig.dev/v1alpha1\n")
				assert.NoError(t, err)
				_, err = f.WriteString("kind: PlatformConfig\n")
				assert.NoError(t, err)
				for _, content := range test.fileConfig {
					assert.NoError(t, err)
					defer os.Remove(f.Name())
					_, err := f.WriteString(content)
					assert.NoError(t, err)
					f.Close()
				}

				content, err := os.ReadFile(f.Name())
				assert.NoError(t, err)
				fmt.Println(string(content))
			}

			if len(test.secretFileConfig) > 0 {
				fs, err = os.CreateTemp("", "rig_secret_test*.yaml")
				assert.NoError(t, err)
				_, err := fs.WriteString("apiVersion: config.rig.dev/v1alpha1\n")
				assert.NoError(t, err)
				_, err = fs.WriteString("kind: PlatformConfig\n")
				assert.NoError(t, err)
				for _, content := range test.secretFileConfig {
					assert.NoError(t, err)
					defer os.Remove(fs.Name())
					_, err := fs.WriteString(content)
					assert.NoError(t, err)
					fs.Close()
				}

				content, err := os.ReadFile(fs.Name())
				assert.NoError(t, err)
				fmt.Println(string(content))
			}

			for k, v := range test.envVars {
				t.Setenv(k, v)
			}

			cfgPath := ""
			if f != nil {
				cfgPath = f.Name()
			}
			secretPath := ""
			if fs != nil {
				secretPath = fs.Name()
			}

			serv, err := NewService(cfgPath, secretPath, manager.NewScheme())
			assert.NoError(t, err)
			c := serv.Get()

			if test.err == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorAs(t, err, &test.err)
			}
			assert.Equal(t, test.expected(), c)
		})
	}
}
