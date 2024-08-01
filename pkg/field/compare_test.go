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
			Name: "add multiple interface",
			From: &platformv1.CapsuleSpec{},
			To: &platformv1.CapsuleSpec{
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
			Changes: []string{
				"Added interface (with port 80)",
				"Added interface (with port 81)",
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
			From: &platformv1.CapsuleSpec{},
			To: &platformv1.CapsuleSpec{
				Env: &platformv1.EnvironmentVariables{
					Raw: map[string]string{
						"key2": "value2",
					},
				},
			},
			Changes: []string{
				"Added env.raw.key2",
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
					Raw: map[string]string{},
				},
			},
			To: &platformv1.CapsuleSpec{
				Env: &platformv1.EnvironmentVariables{
					Raw: map[string]string{
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
				"Removed arg (at index 0)",
				"Removed arg (at index 1)",
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
				"Removed arg (at index 1)",
				"Removed arg (at index 2)",
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
			Name: "add multiple args",
			From: &platformv1.CapsuleSpec{},
			To: &platformv1.CapsuleSpec{
				Args: []string{"arg1", "arg2"},
			},
			Changes: []string{
				"Added arg (at index 0)",
				"Added arg (at index 1)",
			},
		},
		{
			Name: "remove all args",
			From: &platformv1.CapsuleSpec{
				Args: []string{"arg1", "arg2"},
			},
			To: &platformv1.CapsuleSpec{},
			Changes: []string{
				"Removed arg (at index 0)",
				"Removed arg (at index 1)",
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
				"Removed arg (at index 1)",
				"Removed arg (at index 2)",
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
		{
			Name: "Change image",
			From: &platformv1.CapsuleSpec{},
			To: &platformv1.CapsuleSpec{
				Image: "foo/bar:latest",
			},
			Changes: []string{
				"Added image",
			},
		},
		{
			Name: "Change the only arg",
			From: &platformv1.CapsuleSpec{
				Args: []string{"arg1"},
			},
			To: &platformv1.CapsuleSpec{
				Args: []string{"arg2"},
			},
			Changes: []string{
				"Changed arg (at index 0) from 'arg1' to 'arg2'",
			},
		},
		{
			Name: "Change args",
			From: &platformv1.CapsuleSpec{
				Args: []string{"arg1", "arg2"},
			},
			To: &platformv1.CapsuleSpec{
				Args: []string{"arg1", "arg3"},
			},
			Changes: []string{
				"Removed arg (at index 1)",
				"Added arg (at index 1)",
			},
		},
		{
			Name: "change config file",
			From: &platformv1.CapsuleSpec{
				Files: []*platformv1.File{{
					Path:    "/file.yaml",
					String_: "my content 1",
				}},
			},
			To: &platformv1.CapsuleSpec{
				Files: []*platformv1.File{{
					Path:    "/file.yaml",
					String_: "my content 2",
				}},
			},
			Changes: []string{
				"Changed file.string (with path /file.yaml) from 'my content 1' to 'my content 2'",
			},
		},
		{
			Name: "change config file with multiline content",
			From: &platformv1.CapsuleSpec{
				Files: []*platformv1.File{{
					Path:    "/file.yaml",
					String_: "my\ncontent 1",
				}},
			},
			To: &platformv1.CapsuleSpec{
				Files: []*platformv1.File{{
					Path:    "/file.yaml",
					String_: "my\ncontent 2",
				}},
			},
			Changes: []string{
				"Changed file.string (with path /file.yaml)",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			diff, err := Compare(test.From, test.To, "port", "id", "path", "name")
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
			From: &platformv1.CapsuleSpec{},
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
						AsString: "value1",
						Type:     StringType,
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
						AsString: "value2",
						Type:     StringType,
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
						AsString: "value1",
						Type:     StringType,
					},
					Operation: AddedOperation,
				},
				{
					FieldPath: "$.env.raw.key2",
					FieldID:   "$.env.raw.key2",
					To: Value{
						AsString: "value2",
						Type:     StringType,
					},
					Operation: AddedOperation,
				},
			},
		},
		{
			Name: "add one argument",
			From: &platformv1.CapsuleSpec{},
			To: &platformv1.CapsuleSpec{
				Args: []string{"arg1"},
			},
			Changes: []Change{
				{
					FieldPath: "$.args[0]",
					FieldID:   "$.args[0]",
					To: Value{
						AsString: "arg1",
						Type:     StringType,
					},
					Operation: AddedOperation,
				},
			},
		},
		{
			Name: "add one additional argument",
			From: &platformv1.CapsuleSpec{
				Args: []string{
					"arg1",
				},
			},
			To: &platformv1.CapsuleSpec{
				Args: []string{
					"arg1",
					"arg2",
				},
			},
			Changes: []Change{
				{
					FieldPath: "$.args[1]",
					FieldID:   "$.args[1]",
					To: Value{
						AsString: "arg2",
						Type:     StringType,
					},
					Operation: AddedOperation,
				},
			},
		},
		{
			Name: "add one interface from empty",
			From: &platformv1.CapsuleSpec{},
			To: &platformv1.CapsuleSpec{
				Interfaces: []*v1alpha2.CapsuleInterface{
					{
						Name: "foobar",
						Port: 80,
					},
				},
			},
			Changes: []Change{
				{
					FieldPath: "$.interfaces[@port=80]",
					FieldID:   "$.interfaces[@port=80]",
					From: Value{
						AsString: ``,
						Type:     NullType,
					},
					To: Value{
						AsString: "name: foobar\nport: 80\n",
						Type:     MapType,
					},
					Operation: AddedOperation,
				},
			},
		},
		{
			Name: "change route name",
			From: &platformv1.CapsuleSpec{
				Interfaces: []*v1alpha2.CapsuleInterface{
					{
						Name: "foobar",
						Port: 80,
						Routes: []*v1alpha2.HostRoute{
							{
								Id:   "foobarid",
								Host: "example.com",
							},
						},
					},
				},
			},
			To: &platformv1.CapsuleSpec{
				Interfaces: []*v1alpha2.CapsuleInterface{
					{
						Name: "foobar",
						Port: 80,
						Routes: []*v1alpha2.HostRoute{
							{
								Id:   "foobarid",
								Host: "example.org",
							},
						},
					},
				},
			},
			Changes: []Change{
				{
					FieldPath: "$.interfaces[@port=80].routes[@id=foobarid].host",
					FieldID:   "$.interfaces[@port=80].routes[@id=foobarid].host",
					From: Value{
						AsString: "example.com",
						Type:     StringType,
					},
					To: Value{
						AsString: "example.org",
						Type:     StringType,
					},
					Operation: ModifiedOperation,
				},
			},
		},
		{
			Name: "Add route to existing interface",
			From: &platformv1.CapsuleSpec{
				Interfaces: []*v1alpha2.CapsuleInterface{
					{
						Name: "foobar",
						Port: 80,
					},
				},
			},
			To: &platformv1.CapsuleSpec{
				Interfaces: []*v1alpha2.CapsuleInterface{
					{
						Name: "foobar",
						Port: 80,
						Routes: []*v1alpha2.HostRoute{
							{
								Id:   "foobarid",
								Host: "example.org",
							},
						},
					},
				},
			},
			Changes: []Change{
				{
					FieldPath: "$.interfaces[@port=80].routes[@id=foobarid]",
					FieldID:   "$.interfaces[@port=80].routes[@id=foobarid]",
					From: Value{
						AsString: "",
						Type:     NullType,
					},
					To: Value{
						AsString: "id: foobarid\nhost: example.org\n",
						Type:     MapType,
					},
					Operation: AddedOperation,
				},
			},
		},
		{
			Name: "Add interface with route",
			From: &platformv1.CapsuleSpec{},
			To: &platformv1.CapsuleSpec{
				Interfaces: []*v1alpha2.CapsuleInterface{
					{
						Name: "foobar",
						Port: 80,
						Routes: []*v1alpha2.HostRoute{
							{
								Id:   "foobarid",
								Host: "example.org",
							},
						},
					},
				},
			},
			Changes: []Change{
				{
					FieldPath: "$.interfaces[@port=80]",
					FieldID:   "$.interfaces[@port=80]",
					From: Value{
						AsString: "",
						Type:     NullType,
					},
					To: Value{
						AsString: "name: foobar\nport: 80\nroutes:\n    - id: foobarid\n      host: example.org\n",
						Type:     MapType,
					},
					Operation: AddedOperation,
				},
			},
		},
		{
			Name: "Add Horizontal Scale",
			From: &platformv1.CapsuleSpec{},
			To: &platformv1.CapsuleSpec{
				Scale: &platformv1.Scale{
					Horizontal: &platformv1.HorizontalScale{
						Min: 2,
						Max: 5,
						CpuTarget: &v1alpha2.CPUTarget{
							Utilization: 80,
						},
						CustomMetrics: []*v1alpha2.CustomMetric{
							{
								InstanceMetric: &v1alpha2.InstanceMetric{
									MetricName:   "metric1",
									AverageValue: "10",
								},
							},
							{
								InstanceMetric: &v1alpha2.InstanceMetric{
									MetricName:   "metric2",
									AverageValue: "20",
								},
							},
						},
					},
				},
			},
			Changes: []Change{
				{
					FieldPath: "$.scale.horizontal.cpuTarget.utilization",
					FieldID:   "$.scale.horizontal.cpuTarget.utilization",
					From: Value{
						AsString: "",
						Type:     NullType,
					},
					To: Value{
						AsString: "80",
						Type:     OtherType,
					},
					Operation: AddedOperation,
				},
				{
					FieldPath: "$.scale.horizontal.customMetrics[0]",
					FieldID:   "$.scale.horizontal.customMetrics[0]",
					From: Value{
						AsString: "",
						Type:     NullType,
					},
					To: Value{
						AsString: "instanceMetric:\n    averageValue: \"10\"\n    metricName: metric1\n",
						Type:     MapType,
					},
					Operation: AddedOperation,
				},
				{
					FieldPath: "$.scale.horizontal.customMetrics[1]",
					FieldID:   "$.scale.horizontal.customMetrics[1]",
					From: Value{
						AsString: "",
						Type:     NullType,
					},
					To: Value{
						AsString: "instanceMetric:\n    averageValue: \"20\"\n    metricName: metric2\n",
						Type:     MapType,
					},
					Operation: AddedOperation,
				},
				{
					FieldPath: "$.scale.horizontal.max",
					FieldID:   "$.scale.horizontal.max",
					From: Value{
						AsString: "",
						Type:     NullType,
					},
					To: Value{
						AsString: "5",
						Type:     OtherType,
					},
					Operation: AddedOperation,
				},
				{
					FieldPath: "$.scale.horizontal.min",
					FieldID:   "$.scale.horizontal.min",
					From: Value{
						AsString: "",
						Type:     NullType,
					},
					To: Value{
						AsString: "2",
						Type:     OtherType,
					},
					Operation: AddedOperation,
				},
			},
		},
		{
			Name: "Change second custom metrics",
			From: &platformv1.CapsuleSpec{
				Scale: &platformv1.Scale{
					Horizontal: &platformv1.HorizontalScale{
						CustomMetrics: []*v1alpha2.CustomMetric{
							{
								InstanceMetric: &v1alpha2.InstanceMetric{
									MetricName: "metric1",
								},
							},
							{
								InstanceMetric: &v1alpha2.InstanceMetric{
									MetricName: "metric2",
								},
							},
						},
					},
				},
			},
			To: &platformv1.CapsuleSpec{
				Scale: &platformv1.Scale{
					Horizontal: &platformv1.HorizontalScale{
						CustomMetrics: []*v1alpha2.CustomMetric{
							{
								InstanceMetric: &v1alpha2.InstanceMetric{
									MetricName: "metric1",
								},
							},
							{
								InstanceMetric: &v1alpha2.InstanceMetric{
									MetricName: "metric3",
								},
							},
						},
					},
				},
			},
			Changes: []Change{
				{
					FieldPath: "$.scale.horizontal.customMetrics[1]",
					FieldID:   "$.scale.horizontal.customMetrics[1]",
					From: Value{
						AsString: "instanceMetric:\n    metricName: metric2\n",
						Type:     MapType,
					},
					To: Value{
						AsString: "",
						Type:     NullType,
					},
					Operation: RemovedOperation,
				},
				{
					FieldPath: "$.scale.horizontal.customMetrics[1]",
					FieldID:   "$.scale.horizontal.customMetrics[1]",
					From: Value{
						AsString: "",
						Type:     NullType,
					},
					To: Value{
						AsString: "instanceMetric:\n    metricName: metric3\n",
						Type:     MapType,
					},
					Operation: AddedOperation,
				},
			},
		},
		{
			Name: "Change the only arg",
			From: &platformv1.CapsuleSpec{
				Args: []string{"arg1"},
			},
			To: &platformv1.CapsuleSpec{
				Args: []string{"arg2"},
			},
			Changes: []Change{
				{
					FieldPath: "$.args[0]",
					FieldID:   "$.args[0]",
					From: Value{
						AsString: "arg1",
						Type:     StringType,
					},
					To: Value{
						AsString: "arg2",
						Type:     StringType,
					},
					Operation: ModifiedOperation,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			diff, err := Compare(test.From, test.To, SpecKeys...)
			require.NoError(t, err)
			require.Equal(t, test.Changes, diff.Changes)

			applied, err := Apply(test.From, diff.Changes)
			require.NoError(t, err)

			diff2, err := Compare(applied, test.To, SpecKeys...)
			require.NoError(t, err, applied)
			require.Empty(t, diff2.Changes, applied)
		})
	}
}
