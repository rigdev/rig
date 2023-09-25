package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/rigdev/rig/pkg/ptr"
)

func TestDefault(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		spec     CapsuleSpec
		expected CapsuleSpec
	}{
		{
			name:     "replicas default to 1",
			expected: CapsuleSpec{Replicas: ptr.New(int32(1))},
		},
	}

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
				{ConfigMap: &FileContentRef{Name: "test", Key: "test"}},
			},
			expectedErrs: field.ErrorList{
				field.Required(path.Index(0).Child("path"), ""),
			},
		},
		{
			name: "file content ref: name and key are required",
			files: []File{
				{Path: "/test1", ConfigMap: &FileContentRef{}},
				{Path: "/test2", Secret: &FileContentRef{}},
			},
			expectedErrs: field.ErrorList{
				field.Required(path.Index(0).Child("configMap").Child("name"), ""),
				field.Required(path.Index(0).Child("configMap").Child("key"), ""),
				field.Required(path.Index(1).Child("secret").Child("name"), ""),
				field.Required(path.Index(1).Child("secret").Child("key"), ""),
			},
		},
		{
			name: "path should be unique",
			files: []File{
				{Path: "/test", ConfigMap: &FileContentRef{Name: "test", Key: "test"}},
				{Path: "/test", ConfigMap: &FileContentRef{Name: "test", Key: "test"}},
			},
			expectedErrs: field.ErrorList{
				field.Duplicate(path.Index(1).Child("path"), "/test"),
			},
		},
		{
			name: "configMap and secret are mutually exclusive",
			files: []File{
				{
					Path:      "/test",
					ConfigMap: &FileContentRef{Name: "test", Key: "test"},
					Secret:    &FileContentRef{Name: "test", Key: "test"},
				},
			},
			expectedErrs: field.ErrorList{
				field.Invalid(path.Index(0), File{
					Path:      "/test",
					ConfigMap: &FileContentRef{Name: "test", Key: "test"},
					Secret:    &FileContentRef{Name: "test", Key: "test"},
				}, "configMap and secret are mutually exclusive"),
			},
		},
		{
			name: "one of configMap or secret is required",
			files: []File{
				{
					Path: "/test",
				},
			},
			expectedErrs: field.ErrorList{
				field.Required(path.Index(0), "one of configMap or secret is required"),
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
	path := field.NewPath("spec").Child("horizontalScale")
	tests := []struct {
		name         string
		h            HorizontalScale
		expectedErrs field.ErrorList
	}{
		{
			name: "max < min",
			h: HorizontalScale{
				MinReplicas: 10,
				MaxReplicas: 1,
			},
			expectedErrs: []*field.Error{
				field.Invalid(path.Child("maxReplicas"), uint32(1), "maxReplicas cannot be smaller than minReplicas"),
			},
		},
		{
			name: "utilization percentage > 100",
			h: HorizontalScale{
				MinReplicas: 1,
				MaxReplicas: 1,
				CPUTarget: CPUTarget{
					AverageUtilizationPercentage: 110,
				},
			},
			expectedErrs: []*field.Error{
				field.Invalid(
					path.Child("cpuTarget").Child("averageUtilizationPercentage"),
					uint32(110),
					"cannot be larger than 100",
				),
			},
		},
		{
			name: "good, no autoscaling",
			h: HorizontalScale{
				MinReplicas: 10,
			},
		},
		{
			name: "good, with autoscaling",
			h: HorizontalScale{
				MinReplicas: 10,
				MaxReplicas: 30,
				CPUTarget: CPUTarget{
					AverageUtilizationPercentage: 50,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.h.validate(field.NewPath("spec").Child("horizontalScale"))
			assert.Equal(t, tt.expectedErrs, err)
		})
	}
}
