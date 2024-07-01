package field

import (
	"testing"

	platformv1 "github.com/rigdev/rig-go-api/platform/v1"
	"github.com/rigdev/rig-go-api/v1alpha2"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func Test_Apply(t *testing.T) {
	tests := []struct {
		Name     string
		Base     *platformv1.CapsuleSpec
		Changes  []Change
		Expected *platformv1.CapsuleSpec
	}{
		{
			Name: "change interface name",
			Base: &platformv1.CapsuleSpec{
				Interfaces: []*v1alpha2.CapsuleInterface{{
					Name: "foobar",
					Port: 80,
				}},
			},
			Changes: []Change{
				{
					FieldPath: "$.interfaces[@port=80].name",
					To: Value{
						AsString: "foobar-123",
					},
					Operation: ModifiedOperation,
				},
			},
			Expected: &platformv1.CapsuleSpec{
				Interfaces: []*v1alpha2.CapsuleInterface{{
					Name: "foobar-123",
					Port: 80,
				}},
			},
		},
		{
			Name: "remove interface",
			Base: &platformv1.CapsuleSpec{
				Interfaces: []*v1alpha2.CapsuleInterface{{
					Name: "foobar",
					Port: 80,
				}},
			},
			Changes: []Change{
				{
					FieldPath: "$.interfaces[@port=80]",
					Operation: RemovedOperation,
				},
			},
			Expected: &platformv1.CapsuleSpec{
				Interfaces: []*v1alpha2.CapsuleInterface{},
			},
		},
		{
			Name: "add interface",
			Base: &platformv1.CapsuleSpec{},
			Changes: []Change{
				{
					FieldPath: "$.interfaces[@port=80]",
					To: Value{
						AsString: `
- name: foobar
  port: 80`,
					},
					Operation: AddedOperation,
				},
			},
			Expected: &platformv1.CapsuleSpec{
				Interfaces: []*v1alpha2.CapsuleInterface{{
					Name: "foobar",
					Port: 80,
				}},
			},
		},
		{
			Name: "change environment variable",
			Base: &platformv1.CapsuleSpec{
				Env: &platformv1.EnvironmentVariables{
					Raw: map[string]string{
						"key1": "value1",
					},
				},
			},
			Changes: []Change{
				{
					FieldPath: "$.env.raw.key1",
					To: Value{
						AsString: "value2",
					},
					Operation: ModifiedOperation,
				},
			},
			Expected: &platformv1.CapsuleSpec{
				Env: &platformv1.EnvironmentVariables{
					Raw: map[string]string{
						"key1": "value2",
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			result, err := Apply(test.Base, test.Changes)
			require.NoError(t, err)
			d, err := Compare(result, test.Expected)
			require.NoError(t, err)

			require.True(t, proto.Equal(test.Expected, result), d.Changes)
		})
	}
}
