package v1

import (
	"testing"
	"time"

	"github.com/rigdev/rig-go-api/api/v1/capsule"
	v2 "github.com/rigdev/rig-go-api/k8s.io/api/autoscaling/v2"
	"github.com/rigdev/rig-go-api/model"
	platformv1 "github.com/rigdev/rig-go-api/platform/v1"
	"github.com/rigdev/rig-go-api/v1alpha2"
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
		Interfaces: []*v1alpha2.CapsuleInterface{
			{
				Name: "port1",
				Port: 1234,
				Liveness: &v1alpha2.InterfaceProbe{
					Grpc: &v1alpha2.InterfaceGRPCProbe{
						Service: "service",
					},
				},
			},
			{
				Name: "port2",
				Port: 1235,
				Readiness: &v1alpha2.InterfaceProbe{
					Path: "path",
				},
			},
			{
				Name: "port3",
				Port: 1236,
				Liveness: &v1alpha2.InterfaceProbe{
					Tcp: true,
				},
				Routes: []*v1alpha2.HostRoute{{
					Id:   "id",
					Host: "host",
					Paths: []*v1alpha2.HTTPPathRoute{{
						Path:  "path2",
						Match: "Exact",
					}},
					Annotations: map[string]string{
						"key": "value",
					},
				}},
			},
		},
		Scale: &v1alpha2.CapsuleScale{
			Horizontal: &v1alpha2.HorizontalScale{
				Instances: &v1alpha2.Instances{
					Min: 2,
					Max: 5,
				},
				CpuTarget: &v1alpha2.CPUTarget{
					Utilization: 50,
				},
				CustomMetrics: []*v1alpha2.CustomMetric{
					{
						InstanceMetric: &v1alpha2.InstanceMetric{
							MetricName:   "metric",
							MatchLabels:  map[string]string{"label": "value"},
							AverageValue: "5",
						},
					},
					{
						ObjectMetric: &v1alpha2.ObjectMetric{
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
			Vertical: &v1alpha2.VerticalScale{
				Cpu: &v1alpha2.ResourceLimits{
					Request: "0.1",
					Limit:   "0.2",
				},
				Memory: &v1alpha2.ResourceLimits{
					Request: "1000000",
				},
				Gpu: &v1alpha2.ResourceRequest{
					Request: "2",
				},
			},
		},
		Env: &platformv1.EnvironmentVariables{
			Direct: map[string]string{"key1": "value1"},
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

		CronJobs: []*v1alpha2.CronJob{
			{
				Name:     "job",
				Schedule: "* * * * *",
				Command: &v1alpha2.JobCommand{
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
