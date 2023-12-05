package v1alpha2

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v2 "k8s.io/api/autoscaling/v2"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/rigdev/rig/pkg/ptr"
)

func TestDefault(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		spec     CapsuleSpec
		expected CapsuleSpec
	}{}

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			c := &Capsule{
				Spec: test.spec,
			}
			c.Default()
			assert.Equal(t, test.expected, c.Spec)
		})
	}
}

func TestValidateSpec(t *testing.T) {
	t.Parallel()
	specPath := field.NewPath("spec")

	tests := []struct {
		name         string
		spec         CapsuleSpec
		expectedErrs field.ErrorList
	}{
		{
			name: "image is required",
			expectedErrs: field.ErrorList{
				field.Required(specPath.Child("image"), ""),
			},
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := (&Capsule{Spec: test.spec}).validateSpec()
			assert.Equal(t, test.expectedErrs, err)
		})
	}
}

func TestValidateInterfaces(t *testing.T) {
	t.Parallel()
	infsPath := field.NewPath("spec").Child("interfaces")
	tests := []struct {
		name         string
		interfaces   []CapsuleInterface
		expectedErrs field.ErrorList
	}{
		{
			name: "no interfaces returns no errors",
		},
		{
			name:       "name is required",
			interfaces: []CapsuleInterface{{}},
			expectedErrs: field.ErrorList{
				field.Required(infsPath.Index(0).Child("name"), ""),
			},
		},
		{
			name: "names should be unique",
			interfaces: []CapsuleInterface{
				{Name: "test", Port: int32(1)},
				{Name: "test", Port: int32(2)},
			},
			expectedErrs: field.ErrorList{
				field.Duplicate(infsPath.Index(1).Child("name"), "test"),
			},
		},
		{
			name: "ports should be unique",
			interfaces: []CapsuleInterface{
				{Name: "test1", Port: int32(1)},
				{Name: "test2", Port: int32(1)},
			},
			expectedErrs: field.ErrorList{
				field.Duplicate(infsPath.Index(1).Child("port"), int32(1)),
			},
		},
		{
			name: "public: ingress or loadBalancer is required",
			interfaces: []CapsuleInterface{
				{
					Name:   "test",
					Public: &CapsulePublicInterface{},
				},
			},
			expectedErrs: field.ErrorList{
				field.Required(infsPath.Index(0).Child("public"), "ingress or loadBalancer is required"),
			},
		},
		{
			name: "public: ingress and loadBalancer are mutually exclusive",
			interfaces: []CapsuleInterface{
				{
					Name: "test",
					Public: &CapsulePublicInterface{
						Ingress:      &CapsuleInterfaceIngress{Host: "test"},
						LoadBalancer: &CapsuleInterfaceLoadBalancer{},
					},
				},
			},
			expectedErrs: field.ErrorList{
				field.Invalid(
					infsPath.Index(0).Child("public"),
					&CapsulePublicInterface{
						Ingress:      &CapsuleInterfaceIngress{Host: "test"},
						LoadBalancer: &CapsuleInterfaceLoadBalancer{},
					},
					"ingress and loadBalancer are mutually exclusive",
				),
			},
		},
		{
			name: "public: ingress host is required",
			interfaces: []CapsuleInterface{
				{
					Name: "test",
					Public: &CapsulePublicInterface{
						Ingress: &CapsuleInterfaceIngress{},
					},
				},
			},
			expectedErrs: field.ErrorList{
				field.Required(
					infsPath.Index(0).Child("public").Child("ingress").Child("host"), "",
				),
			},
		},
		{
			name: "valid interface probes",
			interfaces: []CapsuleInterface{
				{
					Name:     "test1",
					Port:     1,
					Liveness: &InterfaceProbe{Path: "/health1"},
				},
				{
					Name:      "test2",
					Port:      2,
					Readiness: &InterfaceProbe{Path: "/health2"},
				},
			},
		},
		{
			name: "duplicated interface probes",
			interfaces: []CapsuleInterface{
				{
					Name:      "test1",
					Port:      1,
					Liveness:  &InterfaceProbe{Path: "/health1"},
					Readiness: &InterfaceProbe{Path: "/health1"},
				},
				{
					Name:      "test2",
					Port:      2,
					Liveness:  &InterfaceProbe{Path: "/health2"},
					Readiness: &InterfaceProbe{Path: "/health2"},
				},
			},
			expectedErrs: field.ErrorList{
				field.Duplicate(
					infsPath.Index(1).Child("liveness"), &InterfaceProbe{Path: "/health2"},
				),
				field.Duplicate(
					infsPath.Index(1).Child("readiness"), &InterfaceProbe{Path: "/health2"},
				),
			},
		},
		{
			name: "invalid interface probe",
			interfaces: []CapsuleInterface{
				{
					Name:     "test1",
					Port:     1,
					Liveness: &InterfaceProbe{},
				},
				{
					Name:      "test2",
					Port:      2,
					Readiness: &InterfaceProbe{Path: "health2", TCP: true},
				},
			},
			expectedErrs: field.ErrorList{
				field.Invalid(
					infsPath.Index(0).Child("liveness"),
					&InterfaceProbe{},
					"interface probes must contain one of `path`, `tcp` or `grpc`",
				),
				field.Invalid(
					infsPath.Index(1).Child("readiness").Child("path"), "health2", "path must be an absolute path",
				),
				field.Invalid(
					infsPath.Index(1).Child("readiness"),
					&InterfaceProbe{Path: "health2", TCP: true},
					"interface probes must contain only one of `path`, `tcp` or `grpc`",
				),
			},
		},
	}

	for i := range tests {
		test := tests[i]

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			c := &Capsule{
				Spec: CapsuleSpec{
					Interfaces: test.interfaces,
				},
			}

			_, err := c.validateInterfaces()
			assert.Equal(t, test.expectedErrs, err)
		})
	}
}

func TestValidateEnv(t *testing.T) {
	t.Parallel()
	path := field.NewPath("spec").Child("env").Child("from")
	tests := []struct {
		name         string
		from         []EnvReference
		expectedErrs field.ErrorList
	}{
		{name: "no from should cause no errors"},
		{
			name: "env ref: name and key are required",
			from: []EnvReference{
				{Kind: "ConfigMap"},
				{Kind: "Secret"},
			},
			expectedErrs: field.ErrorList{
				field.Required(path.Index(0).Child("name"), "missing env name"),
				field.Required(path.Index(1).Child("name"), "missing env name"),
			},
		},
		{
			name: "one of configMap or secret is required",
			from: []EnvReference{
				{
					Kind: "", Name: "test",
				},
				{
					Kind: "Deployment", Name: "test",
				},
			},
			expectedErrs: field.ErrorList{
				field.Required(path.Index(0).Child("kind"), "env reference kind is required"),
				field.Invalid(
					path.Index(1).Child("kind"),
					EnvReference{
						Kind: "Deployment",
						Name: "test",
					},
					"env reference kind must be either ConfigMap or Secret"),
			},
		},
	}

	for i := range tests {
		test := tests[i]

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			c := &Capsule{
				Spec: CapsuleSpec{
					Env: &Env{
						From: test.from,
					},
				},
			}

			_, err := c.validateEnv()
			assert.Equal(t, test.expectedErrs, err)
		})
	}
}

func TestValidateFiles(t *testing.T) {
	t.Parallel()
	path := field.NewPath("spec").Child("files")
	tests := []struct {
		name         string
		files        []File
		expectedErrs field.ErrorList
	}{
		{name: "no files should cause no errors"},
		{
			name: "path is required",
			files: []File{
				{Ref: &FileContentReference{Name: "test", Key: "test", Kind: "ConfigMap"}},
			},
			expectedErrs: field.ErrorList{
				field.Required(path.Index(0).Child("path"), ""),
			},
		},
		{
			name: "file content ref: name and key are required",
			files: []File{
				{Path: "/test1", Ref: &FileContentReference{Kind: "ConfigMap"}},
				{Path: "/test2", Ref: &FileContentReference{Kind: "Secret"}},
			},
			expectedErrs: field.ErrorList{
				field.Required(path.Index(0).Child("ref").Child("name"), ""),
				field.Required(path.Index(0).Child("ref").Child("key"), ""),
				field.Required(path.Index(1).Child("ref").Child("name"), ""),
				field.Required(path.Index(1).Child("ref").Child("key"), ""),
			},
		},
		{
			name: "path should be unique",
			files: []File{
				{Path: "/test", Ref: &FileContentReference{Kind: "ConfigMap", Name: "test", Key: "test"}},
				{Path: "/test", Ref: &FileContentReference{Kind: "Secret", Name: "test", Key: "test"}},
			},
			expectedErrs: field.ErrorList{
				field.Duplicate(path.Index(1).Child("path"), "/test"),
			},
		},
		{
			name: "ref is required",
			files: []File{
				{
					Path: "/test",
				},
			},
			expectedErrs: field.ErrorList{
				field.Required(path.Index(0).Child("ref"), "file reference is required"),
			},
		},
		{
			name: "one of configMap or secret is required",
			files: []File{
				{
					Path: "/test1", Ref: &FileContentReference{Kind: "", Name: "test", Key: "test"},
				},
				{
					Path: "/test2", Ref: &FileContentReference{Kind: "Deployment", Name: "test", Key: "test"},
				},
			},
			expectedErrs: field.ErrorList{
				field.Required(path.Index(0).Child("ref").Child("kind"), "file reference kind is required"),
				field.Invalid(
					path.Index(1).Child("ref").Child("kind"),
					File{
						Path: "/test2",
						Ref: &FileContentReference{
							Kind: "Deployment",
							Name: "test",
							Key:  "test",
						},
					},
					"file reference kind must be either ConfigMap or Secret"),
			},
		},
	}

	for i := range tests {
		test := tests[i]

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			c := &Capsule{
				Spec: CapsuleSpec{
					Files: test.files,
				},
			}

			_, err := c.validateFiles()
			assert.Equal(t, test.expectedErrs, err)
		})
	}
}

func Test_HorizontalScaleValidate(t *testing.T) {
	t.Parallel()
	path := field.NewPath("spec").Child("scale").Child("horizontal")
	tests := []struct {
		name         string
		h            HorizontalScale
		expectedErrs field.ErrorList
	}{
		{
			name: "max < min",
			h: HorizontalScale{
				Instances: Instances{
					Min: uint32(10),
					Max: ptr.New(uint32(1)),
				},
			},
			expectedErrs: []*field.Error{
				field.Invalid(path.Child("instances").Child("max"), uint32(1), "max cannot be smaller than min"),
			},
		},
		{
			name: "utilization percentage > 100",
			h: HorizontalScale{
				CPUTarget: &CPUTarget{
					Utilization: ptr.New[uint32](110),
				},
			},
			expectedErrs: []*field.Error{
				field.Invalid(
					path.Child("cpuTarget").Child("utilization"),
					uint32(110),
					"cannot be larger than 100",
				),
			},
		},
		{
			name: "good, no autoscaling",
			h: HorizontalScale{
				Instances: Instances{
					Min: 10,
				},
			},
		},
		{
			name: "good, with autoscaling",
			h: HorizontalScale{
				Instances: Instances{
					Min: 10,
					Max: ptr.New(uint32(30)),
				},
				CPUTarget: &CPUTarget{
					Utilization: ptr.New(uint32(50)),
				},
			},
		},
		{
			name: "both instance and object custom metric",
			h: HorizontalScale{
				Instances: Instances{
					Min: 10,
					Max: ptr.New(uint32(30)),
				},
				CustomMetrics: []CustomMetric{{
					InstanceMetric: &InstanceMetric{
						MetricName:   "metric",
						AverageValue: "1",
					},
					ObjectMetric: &ObjectMetric{
						MetricName:   "metric",
						AverageValue: "1",
						DescribedObject: v2.CrossVersionObjectReference{
							Kind: "Service",
							Name: "service",
						},
					},
				}},
			},
			expectedErrs: []*field.Error{
				field.Invalid(
					path.Child("customMetrics").Index(0),
					CustomMetric{
						InstanceMetric: &InstanceMetric{
							MetricName:   "metric",
							AverageValue: "1",
						},
						ObjectMetric: &ObjectMetric{
							MetricName:   "metric",
							AverageValue: "1",
							DescribedObject: v2.CrossVersionObjectReference{
								Kind: "Service",
								Name: "service",
							},
						},
					},
					"exactly one of 'instanceMetric' and 'objectMetric' must be provided",
				),
			},
		},
		{
			name: "invalid instance metric averageValue",
			h: HorizontalScale{
				Instances: Instances{
					Min: 10,
					Max: ptr.New(uint32(30)),
				},
				CustomMetrics: []CustomMetric{{
					InstanceMetric: &InstanceMetric{
						MetricName:   "metric",
						AverageValue: "p=np",
					},
				}},
			},
			expectedErrs: []*field.Error{
				field.Invalid(
					path.Child("customMetrics").Index(0).Child("instanceMetric").Child("averageValue"),
					"p=np",
					"quantities must match the regular expression '^([+-]?[0-9.]+)([eEinumkKMGTP]*[-+]?[0-9]*)$'",
				),
			},
		},
		{
			name: "invalid object metric, both value and averageValue",
			h: HorizontalScale{
				Instances: Instances{
					Min: 10,
					Max: ptr.New(uint32(30)),
				},
				CustomMetrics: []CustomMetric{{
					ObjectMetric: &ObjectMetric{
						MetricName:   "metric",
						AverageValue: "1",
						Value:        "2",
						DescribedObject: v2.CrossVersionObjectReference{
							Kind: "Service",
							Name: "service",
						},
					},
				}},
			},
			expectedErrs: []*field.Error{
				field.Invalid(
					path.Child("customMetrics").Index(0).Child("objectMetric"),
					&ObjectMetric{
						MetricName:   "metric",
						AverageValue: "1",
						Value:        "2",
						DescribedObject: v2.CrossVersionObjectReference{
							Kind: "Service",
							Name: "service",
						},
					},
					"exactly one of 'value' and 'averageValue' must be provided",
				),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.h.validate(field.NewPath("spec").Child("scale").Child("horizontal"))
			assert.Equal(t, tt.expectedErrs, err)
		})
	}
}
