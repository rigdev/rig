package obj

import (
	"testing"

	"github.com/rigdev/rig/pkg/scheme"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDecodeAny(t *testing.T) {
	str := `
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    rig.dev/owned-by-capsule: my-capsule
  name: my-capsule
  namespace: my-ns
`

	obj, err := DecodeAny([]byte(str), scheme.New())
	require.NoError(t, err)
	require.Equal(t, &corev1.ServiceAccount{
		TypeMeta: v1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      "my-capsule",
			Namespace: "my-ns",
			Labels:    map[string]string{"rig.dev/owned-by-capsule": "my-capsule"},
		},
	}, obj)
}
