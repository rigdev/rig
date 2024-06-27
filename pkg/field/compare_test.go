package field

import (
	"testing"

	platformv1 "github.com/rigdev/rig-go-api/platform/v1"
	"github.com/rigdev/rig-go-api/v1alpha2"
	"github.com/stretchr/testify/require"
)

func Test_Compare(t *testing.T) {
	tests := []struct {
		Name    string
		From    *platformv1.CapsuleSpec
		To      *platformv1.CapsuleSpec
		Changes []string
	}{
		{
			From: &platformv1.CapsuleSpec{
				Interfaces: []*v1alpha2.CapsuleInterface{{
					Name: "foobar",
					Port: 80,
				}},
			},
			To: &platformv1.CapsuleSpec{
				Interfaces: []*v1alpha2.CapsuleInterface{{
					Name: "foobar-123",
					Port: 80,
				}},
			},
			Changes: []string{
				"Changed interface.name (with port 80) from 'foobar' to 'foobar-123'",
			},
		},
		{
			From: &platformv1.CapsuleSpec{},
			To: &platformv1.CapsuleSpec{
				Interfaces: []*v1alpha2.CapsuleInterface{{
					Name: "foobar-123",
					Port: 80,
				}},
			},
			Changes: []string{
				"Added interface (with port 80)",
			},
		},
		{
			From: &platformv1.CapsuleSpec{},
			To: &platformv1.CapsuleSpec{
				Interfaces: []*v1alpha2.CapsuleInterface{{
					Name: "foobar-123",
					Port: 80,
				}},
			},
			Changes: []string{
				"Added interface (with port 80)",
			},
		},
		{
			From: &platformv1.CapsuleSpec{
				Interfaces: []*v1alpha2.CapsuleInterface{{
					Name: "foobar",
					Port: 80,
				}},
			},
			To: &platformv1.CapsuleSpec{
				Interfaces: []*v1alpha2.CapsuleInterface{{
					Name: "foobar",
					Port: 81,
				}},
			},
			Changes: []string{
				"Removed interface (with port 80)",
				"Added interface (with port 81)",
			},
		},
		{
			From: &platformv1.CapsuleSpec{
				Interfaces: []*v1alpha2.CapsuleInterface{{
					Name: "foobar-123",
					Port: 80,
				}},
			},
			To: &platformv1.CapsuleSpec{},
			Changes: []string{
				"Removed interface (with port 80)",
			},
		},
		{
			From: &platformv1.CapsuleSpec{
				Files: []*platformv1.File{
					{
						Path: "/foobar.yaml",
					},
				},
			},
			To: &platformv1.CapsuleSpec{
				Files: []*platformv1.File{
					{
						Path:  "/foobar.yaml",
						Bytes: []byte{0, 1, 2, 3, 4},
					},
				},
			},
			Changes: []string{
				"Added file.bytes (with path /foobar.yaml)",
			},
		},
		{
			From: &platformv1.CapsuleSpec{
				Scale: &platformv1.Scale{
					Horizontal: &platformv1.HorizontalScale{
						Min: 2,
					},
				},
			},
			To: &platformv1.CapsuleSpec{
				Scale: &platformv1.Scale{
					Horizontal: &platformv1.HorizontalScale{
						Min: 3,
					},
				},
			},
			Changes: []string{
				"Changed scale.horizontal.min from '2' to '3'",
			},
		},
		{
			From: &platformv1.CapsuleSpec{
				Env: &platformv1.EnvironmentVariables{
					Direct: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
				},
			},
			To: &platformv1.CapsuleSpec{
				Env: &platformv1.EnvironmentVariables{
					Direct: map[string]string{
						"key1": "value1",
					},
				},
			},
			Changes: []string{
				"Removed env.direct.key2",
			},
		},
		{
			From: &platformv1.CapsuleSpec{
				Env: &platformv1.EnvironmentVariables{
					Direct: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
				},
			},
			To: &platformv1.CapsuleSpec{
				Env: &platformv1.EnvironmentVariables{
					Direct: map[string]string{},
				},
			},
			Changes: []string{
				"Removed env.direct",
			},
		},
		{
			From: &platformv1.CapsuleSpec{
				Env: &platformv1.EnvironmentVariables{
					Direct: map[string]string{
						"key1": "value1",
					},
				},
			},
			To: &platformv1.CapsuleSpec{
				Env: &platformv1.EnvironmentVariables{
					Direct: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
				},
			},
			Changes: []string{
				"Added env.direct.key2",
			},
		},
		{
			From: &platformv1.CapsuleSpec{
				Env: &platformv1.EnvironmentVariables{
					Direct: map[string]string{
						"key1": "value1",
					},
				},
			},
			To: &platformv1.CapsuleSpec{
				Env: &platformv1.EnvironmentVariables{
					Direct: map[string]string{
						"key1": "value1",
						"key2": "value2",
						"key3": "value3",
					},
				},
			},
			Changes: []string{
				"Added env.direct.key2",
				"Added env.direct.key3",
			},
		},
	}

	for _, test := range tests {
		t.Run(t.Name(), func(t *testing.T) {
			diff, err := Compare(test.From, test.To)
			require.NoError(t, err)
			var changes []string
			for _, c := range diff.Changes {
				changes = append(changes, c.String())
			}
			require.Equal(t, test.Changes, changes)
		})
	}
}
