package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name              string
		filePathContent   string
		searchPathContent map[string]string
		envVars           map[string]string
		err               error
		expected          func() Config
	}{
		{
			name:     "nothing set returns default",
			expected: newDefault,
		},
		{
			name: "env overrides default",
			envVars: map[string]string{
				"RIG_PORT": "4242",
			},
			expected: func() Config {
				c := newDefault()
				c.Port = 4242
				return c
			},
		},
		{
			name:            "config is read from file",
			filePathContent: `port: 4242`,
			expected: func() Config {
				c := newDefault()
				c.Port = 4242
				return c
			},
		},
		{
			name: "config is read from search path",
			searchPathContent: map[string]string{
				"/etc/rig": `port: 4242`,
			},
			expected: func() Config {
				c := newDefault()
				c.Port = 4242
				return c
			},
		},
		{
			name: "config from file can partially set in map",
			filePathContent: `repository:
  secret:
    mongodb:
      key: test`,
			expected: func() Config {
				c := newDefault()
				c.Repository.Secret.MongoDB.Key = "test"
				return c
			},
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			var filePath string
			if test.filePathContent != "" {
				f, err := os.CreateTemp("", "rig_test*.yaml")
				assert.NoError(t, err)
				defer os.Remove(f.Name())
				f.WriteString(test.filePathContent)
				f.Close()
				filePath = f.Name()
			}

			for k, v := range test.envVars {
				t.Setenv(k, v)
			}

			var (
				d   string
				err error
			)
			if len(test.searchPathContent) > 0 {
				d, err = os.MkdirTemp("", "rig_test*")
				assert.NoError(t, err)
				defer os.RemoveAll(d)
				for sp, c := range test.searchPathContent {
					tmpSP := filepath.Join(d, sp)
					assert.NoError(t, os.MkdirAll(tmpSP, 0o700))
					assert.NoError(t, os.WriteFile(
						filepath.Join(tmpSP, "server-config.yaml"),
						[]byte(c),
						0o644,
					))
				}
			}

			c, err := new(filePath, mapSlice(keys(test.searchPathContent), func(s string) string {
				return filepath.Join(d, s)
			})...)
			if test.err != nil {
				assert.ErrorAs(t, err, &test.err)
			}
			assert.Equal(t, test.expected(), c)
		})
	}
}

func keys[M ~map[K]V, K comparable, V any](m M) []K {
	keys := make([]K, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func mapSlice[S []V, V any, N any](s S, f func(v V) N) []N {
	ns := make([]N, len(s))
	for i, v := range s {
		ns[i] = f(v)
	}
	return ns
}
