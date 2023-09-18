package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

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
			name: "names should be unique",
			interfaces: []CapsuleInterface{
				{Name: "test", Port: 1},
				{Name: "test", Port: 2},
			},
			expectedErrs: field.ErrorList{
				field.Duplicate(infsPath.Index(1).Child("name"), "test"),
			},
		},
		{
			name: "ports should be unique",
			interfaces: []CapsuleInterface{
				{Name: "test1", Port: 1},
				{Name: "test2", Port: 1},
			},
			expectedErrs: field.ErrorList{
				field.Duplicate(infsPath.Index(1).Child("port"), int32(1)),
			},
		},
		{
			name: "public: ingress or loadBalancer is required",
			interfaces: []CapsuleInterface{
				{Public: &CapsulePublicInterface{}},
			},
			expectedErrs: field.ErrorList{
				field.Required(infsPath.Index(0).Child("public"), "ingress or loadBalancer is required"),
			},
		},
		{
			name: "public: ingress and loadBalancer are mutually exclusive",
			interfaces: []CapsuleInterface{
				{
					Public: &CapsulePublicInterface{
						Ingress:      &CapsuleInterfaceIngress{},
						LoadBalancer: &CapsuleInterfaceLoadBalancer{},
					},
				},
			},
			expectedErrs: field.ErrorList{
				field.Invalid(
					infsPath.Index(0).Child("public"),
					&CapsulePublicInterface{
						Ingress:      &CapsuleInterfaceIngress{},
						LoadBalancer: &CapsuleInterfaceLoadBalancer{},
					},
					"ingress and loadBalancer are mutually exclusive",
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
			name: "path should be unique",
			files: []File{
				{Path: "/test", ConfigMap: &FileContentRef{}},
				{Path: "/test", ConfigMap: &FileContentRef{}},
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
					ConfigMap: &FileContentRef{},
					Secret:    &FileContentRef{},
				},
			},
			expectedErrs: field.ErrorList{
				field.Invalid(path.Index(0), File{
					Path:      "/test",
					ConfigMap: &FileContentRef{},
					Secret:    &FileContentRef{},
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
