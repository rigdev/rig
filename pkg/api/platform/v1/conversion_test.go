package v1

import (
	"testing"
	"time"

	"github.com/rigdev/rig-go-api/api/v1/capsule"
	v2 "github.com/rigdev/rig-go-api/k8s.io/api/autoscaling/v2"
	"github.com/rigdev/rig-go-api/model"
	platformv1 "github.com/rigdev/rig-go-api/platform/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"
)

var (
	rolloutConfig = &capsule.RolloutConfig{
		ImageId: "image",
		Network: &capsule.Network{
			Interfaces: []*capsule.Interface{
				{
					Port: 1234,
					Name: "port1",
					Liveness: &capsule.InterfaceProbe{
						Kind: &capsule.InterfaceProbe_Grpc{
							Grpc: &capsule.InterfaceProbe_GRPC{
								Service: "service",
							},
						},
					},
				},
				{
					Port: 1235,
					Name: "port2",
					Readiness: &capsule.InterfaceProbe{
						Kind: &capsule.InterfaceProbe_Http{
							Http: &capsule.InterfaceProbe_HTTP{
								Path: "path",
							},
						},
					},
				},
				{
					Port: 1236,
					Name: "port3",
					Liveness: &capsule.InterfaceProbe{
						Kind: &capsule.InterfaceProbe_Tcp{
							Tcp: &capsule.InterfaceProbe_TCP{},
						},
					},
					Routes: []*capsule.HostRoute{{
						Host: "host",
						Options: &capsule.RouteOptions{
							Annotations: map[string]string{
								"key": "value",
							},
						},
						Paths: []*capsule.HTTPPathRoute{{
							Path:  "path2",
							Match: capsule.PathMatchType_PATH_MATCH_TYPE_EXACT,
						}},
						Id: "id",
					}},
				},
			},
		},
		ContainerSettings: &capsule.ContainerSettings{
			EnvironmentVariables: map[string]string{
				"key1": "value1",
			},
			Command: "cmd",
			Args:    []string{"arg1", "arg2"},
			Resources: &capsule.Resources{
				Requests: &capsule.ResourceList{
					CpuMillis:   100,
					MemoryBytes: 1_000_000,
				},
				Limits: &capsule.ResourceList{
					CpuMillis: 200,
				},
				GpuLimits: &capsule.GpuLimits{
					// Type:  "gpu", TODO
					Count: 2,
				},
			},
			EnvironmentSources: []*capsule.EnvironmentSource{{
				Name: "some-map",
				Kind: capsule.EnvironmentSource_KIND_CONFIG_MAP,
			}},
		},
		AutoAddRigServiceAccounts: true,
		ConfigFiles: []*capsule.ConfigFile{
			{
				Path:     "/etc/file1.yaml",
				Content:  []byte("hej"),
				IsSecret: false,
			},
			{
				Path:     "/etc/file2.yaml",
				Content:  []byte{0, 0, 0},
				IsSecret: false,
			},
		},
		Replicas: 2,
		HorizontalScale: &capsule.HorizontalScale{
			MaxReplicas: 5,
			MinReplicas: 2,
			CpuTarget: &capsule.CPUTarget{
				AverageUtilizationPercentage: 50,
			},
			CustomMetrics: []*capsule.CustomMetric{
				{
					Metric: &capsule.CustomMetric_Instance{
						Instance: &capsule.InstanceMetric{
							MetricName:   "metric",
							MatchLabels:  map[string]string{"label": "value"},
							AverageValue: "5",
						},
					},
				},
				{
					Metric: &capsule.CustomMetric_Object{
						Object: &capsule.ObjectMetric{
							MetricName:   "metric2",
							MatchLabels:  map[string]string{"label2": "value2"},
							AverageValue: "1",
							Value:        "2",
							ObjectReference: &model.ObjectReference{
								Kind:       "kind",
								Name:       "name",
								ApiVersion: "v1",
							},
						},
					},
				},
			},
		},
		CronJobs: []*capsule.CronJob{{
			JobName:    "job",
			Schedule:   "* * * * *",
			MaxRetries: 1,
			Timeout:    durationpb.New(time.Second * 10),
			JobType: &capsule.CronJob_Command{
				Command: &capsule.JobCommand{
					Command: "cmd",
					Args:    []string{"arg"},
				},
			},
		}},
		Annotations: map[string]string{"annotation": "value"},
	}
	spec = &platformv1.CapsuleSpec{
		Kind:       "CapsuleSpec",
		ApiVersion: "v1",
		Image:      "image",
		Command:    "cmd",
		Args:       []string{"arg1", "arg2"},
		Interfaces: []*platformv1.CapsuleInterface{
			{
				Name: "port1",
				Port: 1234,
				Liveness: &platformv1.InterfaceLivenessProbe{
					Grpc: &platformv1.InterfaceGRPCProbe{
						Service: "service",
					},
				},
			},
			{
				Name: "port2",
				Port: 1235,
				Readiness: &platformv1.InterfaceReadinessProbe{
					Path: "path",
				},
			},
			{
				Name: "port3",
				Port: 1236,
				Liveness: &platformv1.InterfaceLivenessProbe{
					Tcp: true,
				},
				Routes: []*platformv1.HostRoute{{
					Id:   "id",
					Host: "host",
					Paths: []*platformv1.HTTPPathRoute{{
						Path:  "path2",
						Match: "Exact",
					}},
					Annotations: map[string]string{
						"key": "value",
					},
				}},
			},
		},
		Scale: &platformv1.Scale{
			Horizontal: &platformv1.HorizontalScale{
				Min: 2,
				Max: 5,
				CpuTarget: &platformv1.CPUTarget{
					Utilization: 50,
				},
				CustomMetrics: []*platformv1.CustomMetric{
					{
						InstanceMetric: &platformv1.InstanceMetric{
							MetricName:   "metric",
							MatchLabels:  map[string]string{"label": "value"},
							AverageValue: "5",
						},
					},
					{
						ObjectMetric: &platformv1.ObjectMetric{
							MetricName:   "metric2",
							MatchLabels:  map[string]string{"label2": "value2"},
							AverageValue: "1",
							Value:        "2",
							ObjectReference: &v2.CrossVersionObjectReference{
								Kind:       "kind",
								Name:       "name",
								ApiVersion: "v1",
							},
						},
					},
				},
			},
			Vertical: &platformv1.VerticalScale{
				Cpu: &platformv1.ResourceLimits{
					Request: "0.1",
					Limit:   "0.2",
				},
				Memory: &platformv1.ResourceLimits{
					Request: "1000000",
				},
				Gpu: &platformv1.ResourceRequest{
					Request: "2",
				},
			},
		},
		Env: &platformv1.EnvironmentVariables{
			Raw: map[string]string{"key1": "value1"},
			Sources: []*platformv1.EnvironmentSource{{
				Name: "some-map",
				Kind: "ConfigMap",
			}},
		},
		Files: []*platformv1.File{
			{
				Path:     "/etc/file1.yaml",
				String_:  "hej",
				AsSecret: false,
			},
			{
				Path:     "/etc/file2.yaml",
				Bytes:    []byte{0, 0, 0},
				AsSecret: false,
			},
		},

		CronJobs: []*platformv1.CronJob{
			{
				Name:     "job",
				Schedule: "* * * * *",
				Command: &platformv1.JobCommand{
					Command: "cmd",
					Args:    []string{"arg"},
				},
				MaxRetries:     1,
				TimeoutSeconds: 10,
			},
		},
		Annotations:               map[string]string{"annotation": "value"},
		AutoAddRigServiceAccounts: true,
	}
)

func Test_RolloutConfigToCapsuleSpec(t *testing.T) {
	spec2, err := RolloutConfigToCapsuleSpec(rolloutConfig)
	require.NoError(t, err)
	require.Equal(t, spec, spec2)
}

func Test_CapsuleSpecToRolloutConfig(t *testing.T) {
	config, err := CapsuleSpecToRolloutConfig(spec)
	require.NoError(t, err)
	require.Equal(t, rolloutConfig, config)
}

func Test_conversion_both_ways(t *testing.T) {
	config, err := CapsuleSpecToRolloutConfig(spec)
	require.NoError(t, err)
	spec2, err := RolloutConfigToCapsuleSpec(config)
	require.NoError(t, err)
	require.Equal(t, spec, spec2)

	spec2, err = RolloutConfigToCapsuleSpec(rolloutConfig)
	require.NoError(t, err)
	config, err = CapsuleSpecToRolloutConfig(spec2)
	require.NoError(t, err)
	require.Equal(t, rolloutConfig, config)
}

func Test_mergeCapsuleSpec(t *testing.T) {
	tests := []struct {
		name     string
		patch    any
		into     *platformv1.CapsuleSpec
		expected *platformv1.CapsuleSpec
	}{
		{
			name:  "empty projEnv base",
			patch: &platformv1.ProjEnvCapsuleBase{},
			into: &platformv1.CapsuleSpec{
				Kind:       "CapsuleSpec",
				ApiVersion: "v1",
			},
			expected: &platformv1.CapsuleSpec{
				Kind:       "CapsuleSpec",
				ApiVersion: "v1",
			},
		},
		{
			name: "projEnv config files",
			patch: &platformv1.ProjEnvCapsuleBase{
				Files: []*platformv1.File{{
					Path:  "some-path",
					Bytes: []byte{1, 2, 3},
				}},
			},
			into: &platformv1.CapsuleSpec{
				Kind:       "CapsuleSpec",
				ApiVersion: "v1",
				Files: []*platformv1.File{
					{
						Path:  "some-path",
						Bytes: []byte{5, 6, 7},
					},
					{
						Path:  "some-path2",
						Bytes: []byte{1, 2, 3, 4},
					},
				},
			},
			expected: &platformv1.CapsuleSpec{
				Kind:       "CapsuleSpec",
				ApiVersion: "v1",
				Files: []*platformv1.File{
					{
						Path:  "some-path",
						Bytes: []byte{1, 2, 3},
					},
					{
						Path:  "some-path2",
						Bytes: []byte{1, 2, 3, 4},
					},
				},
			},
		},
		{
			name: "projEnv has env vars",
			patch: &platformv1.ProjEnvCapsuleBase{
				Files: []*platformv1.File{},
				Env: &platformv1.EnvironmentVariables{
					Raw: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
				},
			},
			into: &platformv1.CapsuleSpec{
				Kind:       "CapsuleSpec",
				ApiVersion: "v1",
				Env: &platformv1.EnvironmentVariables{
					Raw: map[string]string{
						"key1": "other-value",
						"key3": "value3",
					},
				},
			},
			expected: &platformv1.CapsuleSpec{
				Kind:       "CapsuleSpec",
				ApiVersion: "v1",
				Env: &platformv1.EnvironmentVariables{
					Raw: map[string]string{
						"key1": "value1",
						"key2": "value2",
						"key3": "value3",
					},
				},
			},
		},
		{
			name:  "empty capsule patch",
			patch: &platformv1.CapsuleSpec{},
			into: &platformv1.CapsuleSpec{
				Kind:       "CapsuleSpec",
				ApiVersion: "v1",
				Image:      "image",
				Args:       []string{"arg"},
			},
			expected: &platformv1.CapsuleSpec{
				Kind:       "CapsuleSpec",
				ApiVersion: "v1",
				Image:      "image",
				Args:       []string{"arg"},
			},
		},
		{
			name: "capsule patch with simple values",
			patch: &platformv1.CapsuleSpec{
				Image:       "image",
				Command:     "command",
				Args:        []string{"arg1", "arg2"},
				Annotations: map[string]string{"key2": "value2"},
			},
			into: &platformv1.CapsuleSpec{
				Kind:        "CapsuleSpec",
				ApiVersion:  "v1",
				Image:       "otherimage",
				Command:     "othercommand",
				Args:        []string{"otherarg"},
				Annotations: map[string]string{"key3": "value3"},
			},
			expected: &platformv1.CapsuleSpec{
				Kind:        "CapsuleSpec",
				ApiVersion:  "v1",
				Image:       "image",
				Command:     "command",
				Args:        []string{"arg1", "arg2"},
				Annotations: map[string]string{"key2": "value2", "key3": "value3"},
			},
		},
		{
			name: "interface patch",
			patch: &platformv1.CapsuleSpec{
				Interfaces: []*platformv1.CapsuleInterface{
					{
						Name: "interface1",
						Port: 1001,
						Liveness: &platformv1.InterfaceLivenessProbe{
							Path: "some-path",
							Tcp:  true,
						},
					},
					{
						Name: "interface2",
						Port: 1002,
					},
				},
			},
			into: &platformv1.CapsuleSpec{
				Interfaces: []*platformv1.CapsuleInterface{
					{
						Name: "interface1",
						Port: 1001,
						Readiness: &platformv1.InterfaceReadinessProbe{
							Path: "other-path",
							Tcp:  true,
						},
					},
					{
						Name: "interface3",
						Port: 1003,
					},
				},
			},
			expected: &platformv1.CapsuleSpec{
				Interfaces: []*platformv1.CapsuleInterface{
					{
						Name: "interface1",
						Port: 1001,
						Liveness: &platformv1.InterfaceLivenessProbe{
							Path: "some-path",
							Tcp:  true,
						},
						Readiness: &platformv1.InterfaceReadinessProbe{
							Path: "other-path",
							Tcp:  true,
						},
					},
					{
						Name: "interface2",
						Port: 1002,
					},
					{
						Name: "interface3",
						Port: 1003,
					},
				},
			},
		},
		{
			name: "scale patch",
			patch: &platformv1.CapsuleSpec{
				Scale: &platformv1.Scale{
					Horizontal: &platformv1.HorizontalScale{
						Min: 2,
						Max: 4,
						CustomMetrics: []*platformv1.CustomMetric{
							{
								InstanceMetric: &platformv1.InstanceMetric{
									MetricName:   "some-metric",
									AverageValue: "1",
								},
							},
						},
					},
					Vertical: &platformv1.VerticalScale{
						Cpu: &platformv1.ResourceLimits{
							Request: "1",
						},
					},
				},
			},
			into: &platformv1.CapsuleSpec{
				Scale: &platformv1.Scale{
					Horizontal: &platformv1.HorizontalScale{
						Min: 1,
						Max: 1,
						CustomMetrics: []*platformv1.CustomMetric{
							{
								InstanceMetric: &platformv1.InstanceMetric{
									MetricName:   "some-other-metric",
									AverageValue: "2",
								},
							},
						},
					},
					Vertical: &platformv1.VerticalScale{
						Memory: &platformv1.ResourceLimits{
							Request: "100M",
						},
					},
				},
			},
			expected: &platformv1.CapsuleSpec{
				Scale: &platformv1.Scale{
					Horizontal: &platformv1.HorizontalScale{
						Min: 2,
						Max: 4,
						CustomMetrics: []*platformv1.CustomMetric{
							{
								InstanceMetric: &platformv1.InstanceMetric{
									MetricName:   "some-metric",
									AverageValue: "1",
								},
							},
						},
					},
					Vertical: &platformv1.VerticalScale{
						Cpu: &platformv1.ResourceLimits{
							Request: "1",
						},
						Memory: &platformv1.ResourceLimits{
							Request: "100M",
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := mergeCapsuleSpec(tt.patch, tt.into)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, res)
		})
	}
}
