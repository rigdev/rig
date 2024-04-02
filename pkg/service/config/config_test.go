package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	"github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/obj"
	"github.com/rigdev/rig/pkg/scheme"
)

func TestServiceBuilder(t *testing.T) {
	tests := []struct {
		name    string
		files   [][]byte
		envVars map[string]string
		oCFG    func(*v1alpha1.OperatorConfig) *v1alpha1.OperatorConfig
		pCFG    func(*v1alpha1.PlatformConfig) *v1alpha1.PlatformConfig
		err     error
	}{
		{
			name: "unregistered in scheme",
			files: [][]byte{
				[]byte(`apiVersion: test/v1
kind: Test
`),
			},
			err: &ErrDecoding{},
		},
		{
			name: "unsupported group",
			files: [][]byte{
				[]byte(`apiVersion: v1
kind: Service
`),
			},
			err: &ErrUnsupportedGVK{},
		},
		{
			name: "v1alpha1 operator success",
			files: [][]byte{
				[]byte(`apiVersion: config.rig.dev/v1alpha1
kind: OperatorConfig
devModeEnabled: true
`),
			},
			oCFG: func(cfg *v1alpha1.OperatorConfig) *v1alpha1.OperatorConfig {
				cfg.DevModeEnabled = true
				return cfg
			},
		},
		{
			name: "v1alpha1 platform success",
			files: [][]byte{
				[]byte(`apiVersion: config.rig.dev/v1alpha1
kind: PlatformConfig
port: 42
`),
			},
			pCFG: func(cfg *v1alpha1.PlatformConfig) *v1alpha1.PlatformConfig {
				cfg.Port = 42
				return cfg
			},
		},
	}

	sch := scheme.New()
	ser := obj.NewSerializer(sch)
	decoder := serializer.NewCodecFactory(sch).UniversalDeserializer()

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			dir, err := os.MkdirTemp("", "")
			require.NoError(t, err)
			defer os.RemoveAll(dir)

			fileNames := make([]string, len(test.files))

			for i, file := range test.files {
				f, err := os.CreateTemp(dir, "*.yaml")
				require.NoError(t, err)
				defer f.Close()
				_, err = f.Write(file)
				require.NoError(t, err)
				fileNames[i] = f.Name()
			}

			for k, v := range test.envVars {
				t.Setenv(k, v)
			}

			s, err := newServiceBuilder().
				withDecoder(decoder).
				withFiles(fileNames...).
				withSerializer(ser).
				build()
			if test.err != nil {
				require.ErrorAs(t, err, &test.err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, s)

			if test.oCFG == nil {
				test.oCFG = func(cfg *v1alpha1.OperatorConfig) *v1alpha1.OperatorConfig {
					return cfg
				}
			}
			if test.pCFG == nil {
				test.pCFG = func(cfg *v1alpha1.PlatformConfig) *v1alpha1.PlatformConfig {
					return cfg
				}
			}

			defaultOCFG := (&v1alpha1.OperatorConfig{}).Default()
			defaultPCFG := v1alpha1.NewDefaultPlatform()

			assert.Equal(t, test.oCFG(defaultOCFG), s.Operator())
			assert.Equal(t, test.pCFG(defaultPCFG), s.Platform())
		})
	}
}
