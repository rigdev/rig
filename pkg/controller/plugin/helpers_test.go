package plugin

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/go-logr/logr"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/pipeline"
	"github.com/rigdev/rig/pkg/scheme"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_ParseCapsuleTemplatedConfig(t *testing.T) {
	name, namespace := "name", "namespace"
	vm := scheme.NewVersionMapperFromScheme(scheme.New())
	p := pipeline.NewCapsulePipeline(nil, scheme.New(), vm, logr.FromContextOrDiscard(context.Background()))

	req := pipeline.NewCapsuleRequest(p, &v1alpha2.Capsule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha2.CapsuleSpec{
			Extensions: map[string]json.RawMessage{
				"ext": json.RawMessage(`{"field": "value"}`),
			},
			Scale: v1alpha2.CapsuleScale{
				Horizontal: v1alpha2.HorizontalScale{
					Instances: v1alpha2.Instances{
						Min: 69,
					},
				},
			},
		},
	}, nil)

	s := `hej: asdf
hej2: {{ .capsuleExtensions.ext.field  }}
hej3: {{ .capsule.spec.scale.horizontal.instances.min }}`
	conf, err := ParseCapsuleTemplatedConfig[config]([]byte(s), req)
	require.NoError(t, err)

	require.Equal(t, config{
		Hej:  "asdf",
		Hej2: "value",
		Hej3: 69,
	}, conf)
}

type config struct {
	Hej  string `json:"hej"`
	Hej2 string `json:"hej2"`
	Hej3 int    `json:"hej3"`
}
