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
			Name: "change interface name",
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
			Name: "add interface",
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
			Name: "change port",
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
			Name: "remove interface",
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
			Name: "remove multiple interfaces",
			From: &platformv1.CapsuleSpec{
				Interfaces: []*v1alpha2.CapsuleInterface{
					{
						Name: "foobar-123",
						Port: 80,
					},
					{
						Name: "foobar-456",
						Port: 81,
					},
				},
			},
			To: &platformv1.CapsuleSpec{},
			Changes: []string{
				"Removed interface (with port 80)",
				"Removed interface (with port 81)",
			},
		},
		{
			Name: "add bytes to file",
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
			Name: "change scale replicas",
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
			Name: "remove one environment variable",
			From: &platformv1.CapsuleSpec{
				Env: &platformv1.EnvironmentVariables{
					Raw: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
				},
			},
			To: &platformv1.CapsuleSpec{
				Env: &platformv1.EnvironmentVariables{
					Raw: map[string]string{
						"key1": "value1",
					},
				},
			},
			Changes: []string{
				"Removed env.raw.key2",
			},
		},
		{
			Name: "remove all envs",
			From: &platformv1.CapsuleSpec{
				Env: &platformv1.EnvironmentVariables{
					Raw: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
				},
			},
			To: &platformv1.CapsuleSpec{
				Env: &platformv1.EnvironmentVariables{
					Raw: map[string]string{},
				},
			},
			Changes: []string{
				"Removed env.raw.key1",
				"Removed env.raw.key2",
			},
		},
		{
			Name: "add one environment variable",
			From: &platformv1.CapsuleSpec{
				Env: &platformv1.EnvironmentVariables{
					Raw: map[string]string{
						"key1": "value1",
					},
				},
			},
			To: &platformv1.CapsuleSpec{
				Env: &platformv1.EnvironmentVariables{
					Raw: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
				},
			},
			Changes: []string{
				"Added env.raw.key2",
			},
		},
		{
			Name: "add multiple environment variables",
			From: &platformv1.CapsuleSpec{
				Env: &platformv1.EnvironmentVariables{
					Raw: map[string]string{
						"key1": "value1",
					},
				},
			},
			To: &platformv1.CapsuleSpec{
				Env: &platformv1.EnvironmentVariables{
					Raw: map[string]string{
						"key1": "value1",
						"key2": "value2",
						"key3": "value3",
						"key4": "value4",
					},
				},
			},
			Changes: []string{
				"Added env.raw.key2",
				"Added env.raw.key3",
				"Added env.raw.key4",
			},
		},
		{
			Name: "remove all args",
			From: &platformv1.CapsuleSpec{
				Args: []string{"arg1", "arg2"},
			},
			To: &platformv1.CapsuleSpec{},
			Changes: []string{
				"Removed args.arg1",
				"Removed args.arg2",
			},
		},
		{
			Name: "remove multiple args",
			From: &platformv1.CapsuleSpec{
				Args: []string{"arg1", "arg2", "arg3"},
			},
			To: &platformv1.CapsuleSpec{
				Args: []string{"arg1"},
			},
			Changes: []string{
				"Removed args.arg2",
				"Removed args.arg3",
			},
		},
		{
			Name: "Remove all cronjobs",
			From: &platformv1.CapsuleSpec{
				CronJobs: []*v1alpha2.CronJob{
					{
						Name: "cronjob1",
					},
					{
						Name: "cronjob2",
					},
				},
			},
			To: &platformv1.CapsuleSpec{},
			Changes: []string{
				"Removed cronJob (with name cronjob1)",
				"Removed cronJob (with name cronjob2)",
			},
		},
		{
			Name: "Add a cronjob",
			From: &platformv1.CapsuleSpec{},
			To: &platformv1.CapsuleSpec{
				CronJobs: []*v1alpha2.CronJob{
					{
						Name: "cronjob1",
					},
				},
			},
			Changes: []string{
				"Added cronJob (with name cronjob1)",
			},
		},
		{
			Name: "modify a cronjob",
			From: &platformv1.CapsuleSpec{
				CronJobs: []*v1alpha2.CronJob{
					{
						Name:     "cronjob1",
						Schedule: "0 0 * * *",
					},
				},
			},
			To: &platformv1.CapsuleSpec{
				CronJobs: []*v1alpha2.CronJob{
					{
						Name:     "cronjob1",
						Schedule: "0 5 * * *",
					},
				},
			},
			Changes: []string{
				"Changed cronJob.schedule (with name cronjob1) from '0 0 * * *' to '0 5 * * *'",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
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

func Test_Compare_Exact(t *testing.T) {
	tests := []struct {
		Name    string
		From    *platformv1.CapsuleSpec
		To      *platformv1.CapsuleSpec
		Changes []Change
	}{
		{
			Name: "add one environment variable",
			From: &platformv1.CapsuleSpec{
				Env: &platformv1.EnvironmentVariables{
					Raw: map[string]string{},
				},
			},
			To: &platformv1.CapsuleSpec{
				Env: &platformv1.EnvironmentVariables{
					Raw: map[string]string{
						"key1": "value1",
					},
				},
			},
			Changes: []Change{
				{
					FieldPath: "$.env.raw.key1",
					FieldID:   "$.env.raw.key1",
					To: Value{
						AsString: "key1: value1\n",
						Type:     MapType,
					},
					Operation: AddedOperation,
				},
			},
		},
		{
			Name: "add one additional environment variable",
			From: &platformv1.CapsuleSpec{
				Env: &platformv1.EnvironmentVariables{
					Raw: map[string]string{
						"key1": "value1",
					},
				},
			},
			To: &platformv1.CapsuleSpec{
				Env: &platformv1.EnvironmentVariables{
					Raw: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
				},
			},
			Changes: []Change{
				{
					FieldPath: "$.env.raw.key2",
					FieldID:   "$.env.raw.key2",
					To: Value{
						AsString: "key2: value2\n",
						Type:     MapType,
					},
					Operation: AddedOperation,
				},
			},
		},
		{
			Name: "add multiple environment variables",
			From: &platformv1.CapsuleSpec{
				Env: &platformv1.EnvironmentVariables{
					Raw: map[string]string{},
				},
			},
			To: &platformv1.CapsuleSpec{
				Env: &platformv1.EnvironmentVariables{
					Raw: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
				},
			},
			Changes: []Change{
				{
					FieldPath: "$.env.raw.key1",
					FieldID:   "$.env.raw.key1",
					To: Value{
						AsString: "key1: value1\n",
						Type:     MapType,
					},
					Operation: AddedOperation,
				},
				{
					FieldPath: "$.env.raw.key2",
					FieldID:   "$.env.raw.key2",
					To: Value{
						AsString: "key2: value2\n",
						Type:     MapType,
					},
					Operation: AddedOperation,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			diff, err := Compare(test.From, test.To)
			require.NoError(t, err)
			require.Equal(t, test.Changes, diff.Changes)
		})
	}
}
