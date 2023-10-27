package hash_test

import (
	"hash/crc32"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	righash "github.com/rigdev/rig/pkg/hash"
)

func TestSecret(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		secretName      string
		secretData      map[string][]byte
		expectedWritten string
	}{
		{
			name:            "secret name is included in hash",
			secretName:      "test",
			expectedWritten: "test",
		},
		{
			name: "keys are sorted",
			secretData: map[string][]byte{
				"c": []byte("C"),
				"b": []byte("B"),
				"d": []byte("D"),
				"a": []byte("A"),
			},
			expectedWritten: "aAbBcCdD",
		},
		{
			name:       "secret name and data are both included in hash",
			secretName: "test",
			secretData: map[string][]byte{
				"a": []byte("A"),
			},
			expectedWritten: "testaA",
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			expectedH := crc32.NewIEEE()
			_, err := expectedH.Write([]byte(test.expectedWritten))
			require.NoError(t, err)

			s := &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: test.secretName},
				Data:       test.secretData,
			}

			h := crc32.NewIEEE()
			assert.NoError(t, righash.Secret(h, s))
			assert.Equal(t, expectedH.Sum32(), h.Sum32())
		})
	}
}

func TestConfigMap(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		cmName          string
		cmBinaryData    map[string][]byte
		cmData          map[string]string
		expectedWritten string
	}{
		{
			name:            "configmap name is included in hash",
			cmName:          "test",
			expectedWritten: "test",
		},
		{
			name: "keys are sorted",
			cmData: map[string]string{
				"c": "C",
				"b": "B",
				"d": "D",
				"a": "A",
			},
			expectedWritten: "aAbBcCdD",
		},
		{
			name:            "binary data keys are sorted",
			expectedWritten: "aAbBcCdD",
			cmBinaryData: map[string][]byte{
				"c": []byte("C"),
				"b": []byte("B"),
				"d": []byte("D"),
				"a": []byte("A"),
			},
		},
		{
			name:   "configmap name and data and binary data all included in hash",
			cmName: "test",
			cmData: map[string]string{
				"a": "A",
			},
			cmBinaryData: map[string][]byte{
				"b": []byte("B"),
			},
			expectedWritten: "testaAbB",
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			expectedH := crc32.NewIEEE()
			_, err := expectedH.Write([]byte(test.expectedWritten))
			require.NoError(t, err)

			cm := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: test.cmName},
				Data:       test.cmData,
				BinaryData: test.cmBinaryData,
			}

			h := crc32.NewIEEE()
			assert.NoError(t, righash.ConfigMap(h, cm))
			assert.Equal(t, expectedH.Sum32(), h.Sum32())
		})
	}
}

func TestSecretKeys(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		secretName      string
		secretData      map[string][]byte
		referencedKeys  []string
		expectedWritten string
	}{
		{
			name:            "secret name is included in hash",
			secretName:      "test",
			expectedWritten: "test",
		},
		{
			name:           "only referenced keys are sorted",
			referencedKeys: []string{"a", "b"},
			secretData: map[string][]byte{
				"c": []byte("C"),
				"b": []byte("B"),
				"d": []byte("D"),
				"a": []byte("A"),
			},
			expectedWritten: "aAbB",
		},
		{
			name:           "secret name and data are both included in hash",
			referencedKeys: []string{"a"},
			secretName:     "test",
			secretData: map[string][]byte{
				"a": []byte("A"),
			},
			expectedWritten: "testaA",
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			expectedH := crc32.NewIEEE()
			_, err := expectedH.Write([]byte(test.expectedWritten))
			require.NoError(t, err)

			s := &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: test.secretName},
				Data:       test.secretData,
			}

			h := crc32.NewIEEE()
			assert.NoError(t, righash.SecretKeys(h, test.referencedKeys, s))
			assert.Equal(t, expectedH.Sum32(), h.Sum32())
		})
	}
}

func TestConfigMapKeys(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		cmName          string
		cmData          map[string]string
		cmBinaryData    map[string][]byte
		referencedKeys  []string
		expectedWritten string
	}{
		{
			name:            "secret name is included in hash",
			cmName:          "test",
			expectedWritten: "test",
		},
		{
			name:           "only referenced keys are sorted",
			referencedKeys: []string{"a", "b"},
			cmBinaryData: map[string][]byte{
				"c": []byte("C"),
				"b": []byte("B"),
				"d": []byte("D"),
				"a": []byte("A"),
			},
			expectedWritten: "aAbB",
		},
		{
			name:           "secret name and data and binary data are all included in hash",
			referencedKeys: []string{"a", "b"},
			cmName:         "test",
			cmData: map[string]string{
				"b": "B",
			},
			cmBinaryData: map[string][]byte{
				"a": []byte("A"),
			},
			expectedWritten: "testbBaA",
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			expectedH := crc32.NewIEEE()
			_, err := expectedH.Write([]byte(test.expectedWritten))
			require.NoError(t, err)

			cm := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: test.cmName},
				Data:       test.cmData,
				BinaryData: test.cmBinaryData,
			}

			h := crc32.NewIEEE()
			assert.NoError(t, righash.ConfigMapKeys(h, test.referencedKeys, cm))
			assert.Equal(t, expectedH.Sum32(), h.Sum32())
		})
	}
}
