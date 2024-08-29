package field

import (
	"testing"

	platformv1 "github.com/rigdev/rig-go-api/platform/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

type test struct {
	Name     string
	Base     *platformv1.CapsuleSpec
	Changes  []Change
	Expected *platformv1.CapsuleSpec
}

func Test_Apply_Interfaces(t *testing.T) {
	tests := []test{
		{
			Name: "change interface port",
			Base: &platformv1.CapsuleSpec{
				Interfaces: []*platformv1.CapsuleInterface{{
					Name: "foobar",
					Port: 80,
				}},
			},
			Changes: []Change{
				{
					FieldPath: "$.interfaces[@port=80].port",
					To: Value{
						AsString: "81",
					},
					Operation: ModifiedOperation,
				},
			},
			Expected: &platformv1.CapsuleSpec{
				Interfaces: []*platformv1.CapsuleInterface{{
					Name: "foobar",
					Port: 81,
				}},
			},
		},
		{
			Name: "change interface name",
			Base: &platformv1.CapsuleSpec{
				Interfaces: []*platformv1.CapsuleInterface{{
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
				Interfaces: []*platformv1.CapsuleInterface{{
					Name: "foobar-123",
					Port: 80,
				}},
			},
		},
		{
			Name: "remove interface",
			Base: &platformv1.CapsuleSpec{
				Interfaces: []*platformv1.CapsuleInterface{{
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
				Interfaces: []*platformv1.CapsuleInterface{},
			},
		},
		{
			Name: "add interface",
			Base: &platformv1.CapsuleSpec{},
			Changes: []Change{
				{
					FieldPath: "$.interfaces[@port=80]",
					To: Value{
						AsString: "name: foobar\nport: 80\n",
					},
					Operation: AddedOperation,
				},
			},
			Expected: &platformv1.CapsuleSpec{
				Interfaces: []*platformv1.CapsuleInterface{{
					Name: "foobar",
					Port: 80,
				}},
			},
		},
		{
			Name: "add multiple interfaces",
			Base: &platformv1.CapsuleSpec{},
			Changes: []Change{
				{
					FieldPath: "$.interfaces[@port=80]",
					To: Value{
						AsString: "name: foobar\nport: 80\n",
					},
					Operation: AddedOperation,
				},
				{
					FieldPath: "$.interfaces[@port=81]",
					To: Value{
						AsString: "name: foobar2\nport: 81\n",
					},
					Operation: AddedOperation,
				},
			},
			Expected: &platformv1.CapsuleSpec{
				Interfaces: []*platformv1.CapsuleInterface{
					{
						Name: "foobar",
						Port: 80,
					},
					{
						Name: "foobar2",
						Port: 81,
					},
				},
			},
		},
	}

	runTests(t, tests)
}

func Test_Apply_Routes(t *testing.T) {
	tests := []test{
		{
			Name: "change route host",
			Base: &platformv1.CapsuleSpec{
				Interfaces: []*platformv1.CapsuleInterface{
					{
						Name: "foobar",
						Port: 80,
						Routes: []*platformv1.HostRoute{
							{
								Id:   "foobarid",
								Host: "example.com",
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
						AsString: "example2.com",
						Type:     StringType,
					},
					Operation: ModifiedOperation,
				},
			},
			Expected: &platformv1.CapsuleSpec{
				Interfaces: []*platformv1.CapsuleInterface{
					{
						Name: "foobar",
						Port: 80,
						Routes: []*platformv1.HostRoute{
							{
								Id:   "foobarid",
								Host: "example2.com",
							},
						},
					},
				},
			},
		},
		{
			Name: "Remove route",
			Base: &platformv1.CapsuleSpec{
				Interfaces: []*platformv1.CapsuleInterface{
					{
						Name: "foobar",
						Port: 80,
						Routes: []*platformv1.HostRoute{
							{
								Id:   "foobarid",
								Host: "example.com",
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
						AsString: "example.com",
						Type:     StringType,
					},
					To: Value{
						AsString: "",
						Type:     NullType,
					},
					Operation: RemovedOperation,
				},
			},
			Expected: &platformv1.CapsuleSpec{
				Interfaces: []*platformv1.CapsuleInterface{
					{
						Name:   "foobar",
						Port:   80,
						Routes: []*platformv1.HostRoute{},
					},
				},
			},
		},
		{
			Name: "Add Route to existing interface",
			Base: &platformv1.CapsuleSpec{
				Interfaces: []*platformv1.CapsuleInterface{
					{
						Name: "foobar",
						Port: 80,
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
						AsString: "id: foobarid\nhost: example.com\n",
						Type:     MapType,
					},
					Operation: AddedOperation,
				},
			},
			Expected: &platformv1.CapsuleSpec{
				Interfaces: []*platformv1.CapsuleInterface{
					{
						Name: "foobar",
						Port: 80,
						Routes: []*platformv1.HostRoute{
							{
								Id:   "foobarid",
								Host: "example.com",
							},
						},
					},
				},
			},
		},
	}

	runTests(t, tests)
}

func Test_Apply_Env_Vars(t *testing.T) {
	tests := []test{
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
		{
			Name: "add environment variable",
			Base: &platformv1.CapsuleSpec{},
			Changes: []Change{
				{
					FieldPath: "$.env.raw.key1",
					To: Value{
						AsString: "value1",
						Type:     StringType,
					},
					Operation: AddedOperation,
				},
			},
			Expected: &platformv1.CapsuleSpec{
				Env: &platformv1.EnvironmentVariables{
					Raw: map[string]string{
						"key1": "value1",
					},
				},
			},
		},
		{
			Name: "add multiple environment variables",
			Base: &platformv1.CapsuleSpec{},
			Changes: []Change{
				{
					FieldPath: "$.env.raw.key1",
					To: Value{
						AsString: "value1",
					},
					Operation: AddedOperation,
				},
				{
					FieldPath: "$.env.raw.key2",
					To: Value{
						AsString: "value2",
					},
					Operation: AddedOperation,
				},
			},
			Expected: &platformv1.CapsuleSpec{
				Env: &platformv1.EnvironmentVariables{
					Raw: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
				},
			},
		},
		{
			Name: "remove environment variable",
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
					Operation: RemovedOperation,
				},
			},
			Expected: &platformv1.CapsuleSpec{
				Env: &platformv1.EnvironmentVariables{
					Raw: map[string]string{},
				},
			},
		},
	}

	runTests(t, tests)
}

func Test_Apply_Primary_Fields(t *testing.T) {
	tests := []test{
		{
			Name: "Change Image",
			Base: &platformv1.CapsuleSpec{
				Image: "foo/bar:latest",
			},
			Changes: []Change{
				{
					FieldPath: "$.image",
					To: Value{
						AsString: "foo/bar:1.0.0",
					},
					Operation: ModifiedOperation,
				},
			},
			Expected: &platformv1.CapsuleSpec{
				Image: "foo/bar:1.0.0",
			},
		},
		{
			Name: "Add image",
			Base: &platformv1.CapsuleSpec{},
			Changes: []Change{
				{
					FieldPath: "$.image",
					To: Value{
						AsString: "foo/bar:latest",
					},
					Operation: AddedOperation,
				},
			},
			Expected: &platformv1.CapsuleSpec{
				Image: "foo/bar:latest",
			},
		},
		{
			Name: "Remove image",
			Base: &platformv1.CapsuleSpec{
				Image: "foo/bar:latest",
			},
			Changes: []Change{
				{
					FieldPath: "$.image",
					Operation: RemovedOperation,
				},
			},
			Expected: &platformv1.CapsuleSpec{},
		},
		{
			Name: "Change Command",
			Base: &platformv1.CapsuleSpec{
				Command: "echo",
			},
			Changes: []Change{
				{
					FieldPath: "$.command",
					To: Value{
						AsString: "ls",
					},
					Operation: ModifiedOperation,
				},
			},
			Expected: &platformv1.CapsuleSpec{
				Command: "ls",
			},
		},
	}

	runTests(t, tests)
}

func Test_Apply_Args(t *testing.T) {
	tests := []test{
		{
			Name: "Change Args",
			Base: &platformv1.CapsuleSpec{
				Args: []string{"arg1"},
			},
			Changes: []Change{
				{
					FieldPath: "$.args[0]",
					From: Value{
						AsString: "arg1",
					},
					To: Value{
						AsString: "arg2",
					},
					Operation: ModifiedOperation,
				},
			},
			Expected: &platformv1.CapsuleSpec{
				Args: []string{"arg2"},
			},
		},
		{
			Name: "Add Args",
			Base: &platformv1.CapsuleSpec{},
			Changes: []Change{
				{
					FieldPath: "$.args.arg1",
					To: Value{
						AsString: `arg1`,
					},
					Operation: AddedOperation,
				},
				{
					FieldPath: "args.arg2",
					To: Value{
						AsString: `arg2`,
					},
					Operation: AddedOperation,
				},
			},
			Expected: &platformv1.CapsuleSpec{
				Args: []string{"arg1", "arg2"},
			},
		},
	}

	runTests(t, tests)
}

func Test_Apply_Scale(t *testing.T) {
	tests := []test{
		{
			Name: "Change Scale",
			Base: &platformv1.CapsuleSpec{
				Scale: &platformv1.Scale{
					Horizontal: &platformv1.HorizontalScale{
						Min: 1,
					},
				},
			},
			Changes: []Change{
				{
					FieldPath: "$.scale.horizontal.min",
					To: Value{
						AsString: "2",
						Type:     StringType,
					},
					Operation: ModifiedOperation,
				},
			},
			Expected: &platformv1.CapsuleSpec{
				Scale: &platformv1.Scale{
					Horizontal: &platformv1.HorizontalScale{
						Min: 2,
					},
				},
			},
		},
		{
			Name: "Add Horizontal Autoscaler",
			Base: &platformv1.CapsuleSpec{
				Scale: &platformv1.Scale{},
			},
			Expected: &platformv1.CapsuleSpec{
				Scale: &platformv1.Scale{
					Horizontal: &platformv1.HorizontalScale{
						Min: 2,
						Max: 10,
						CpuTarget: &platformv1.CPUTarget{
							Utilization: 80,
						},
						CustomMetrics: []*platformv1.CustomMetric{
							{
								InstanceMetric: &platformv1.InstanceMetric{
									MetricName:   "metric1",
									AverageValue: "10",
								},
							},
						},
					},
				},
			},
			Changes: []Change{
				{
					FieldPath: "$.scale.horizontal.min",
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
				{
					FieldPath: "$.scale.horizontal.max",
					From: Value{
						AsString: "",
						Type:     NullType,
					},
					To: Value{
						AsString: "10",
						Type:     OtherType,
					},
					Operation: AddedOperation,
				},
				{
					FieldPath: "$.scale.horizontal.cpuTarget.utilization",
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
					From: Value{
						AsString: "",
						Type:     NullType,
					},
					To: Value{
						AsString: "instanceMetric:\n  metricName: metric1\n  averageValue: 10\n",
						Type:     OtherType,
					},
					Operation: AddedOperation,
				},
			},
		},
		{
			Name: "Change second custom metrics",
			Base: &platformv1.CapsuleSpec{
				Scale: &platformv1.Scale{
					Horizontal: &platformv1.HorizontalScale{
						CustomMetrics: []*platformv1.CustomMetric{
							{
								InstanceMetric: &platformv1.InstanceMetric{
									MetricName: "metric1",
								},
							},
							{
								InstanceMetric: &platformv1.InstanceMetric{
									MetricName: "metric2",
								},
							},
						},
					},
				},
			},
			Expected: &platformv1.CapsuleSpec{
				Scale: &platformv1.Scale{
					Horizontal: &platformv1.HorizontalScale{
						CustomMetrics: []*platformv1.CustomMetric{
							{
								InstanceMetric: &platformv1.InstanceMetric{
									MetricName: "metric1",
								},
							},
							{
								InstanceMetric: &platformv1.InstanceMetric{
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
	}

	runTests(t, tests)
}

func Test_Run_All(t *testing.T) {
	t.Run("Interfaces", Test_Apply_Interfaces)
	t.Run("Routes", Test_Apply_Routes)
	t.Run("Env Vars", Test_Apply_Env_Vars)
	t.Run("Primary Fields", Test_Apply_Primary_Fields)
	t.Run("Args", Test_Apply_Args)
	t.Run("Scale", Test_Apply_Scale)
}

func runTests(t *testing.T, tests []test) {
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			result, err := Apply(test.Base, test.Changes)
			spec := result.(*platformv1.CapsuleSpec)
			require.NoError(t, err)
			d, err := Compare(spec, test.Expected, SpecKeys...)
			require.Empty(t, d.Changes)
			require.NoError(t, err)
			require.True(t, proto.Equal(test.Expected, result), d.Changes)
		})
	}
}
