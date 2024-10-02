package v1

import (
	"testing"

	platformv1 "github.com/rigdev/rig-go-api/platform/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func Test_initialise(t *testing.T) {
	tests := []struct {
		name     string
		input    proto.Message
		expected proto.Message
	}{
		{
			name:  "empty capsule",
			input: &platformv1.Capsule{},
			expected: &platformv1.Capsule{
				Spec: &platformv1.CapsuleSpec{
					Annotations: map[string]string{},
					Env: &platformv1.EnvironmentVariables{
						Raw: map[string]string{},
					},
					Scale: &platformv1.Scale{
						Horizontal: &platformv1.HorizontalScale{
							Instances: &platformv1.Instances{},
							CpuTarget: &platformv1.CPUTarget{},
						},
						Vertical: &platformv1.VerticalScale{
							Cpu:    &platformv1.ResourceLimits{},
							Memory: &platformv1.ResourceLimits{},
							Gpu:    &platformv1.ResourceRequest{},
						},
					},
					Extensions: map[string]*structpb.Struct{},
				},
			},
		},
		{
			name: "empty capsule",
			input: &platformv1.Capsule{
				Kind:        "Capsule",
				ApiVersion:  "platform.rig.dev",
				Name:        "capsule",
				Project:     "project",
				Environment: "environment",
				Spec: &platformv1.CapsuleSpec{
					Annotations: map[string]string{
						"key": "value",
					},
					Image: "image",
					Env: &platformv1.EnvironmentVariables{
						Raw: map[string]string{
							"field": "asdf",
						},
					},
				},
			},
			expected: &platformv1.Capsule{
				Kind:        "Capsule",
				ApiVersion:  "platform.rig.dev",
				Name:        "capsule",
				Project:     "project",
				Environment: "environment",
				Spec: &platformv1.CapsuleSpec{
					Annotations: map[string]string{
						"key": "value",
					},
					Image: "image",
					Env: &platformv1.EnvironmentVariables{
						Raw: map[string]string{
							"field": "asdf",
						},
					},
					Scale: &platformv1.Scale{
						Horizontal: &platformv1.HorizontalScale{
							Instances: &platformv1.Instances{},
							CpuTarget: &platformv1.CPUTarget{},
						},
						Vertical: &platformv1.VerticalScale{
							Cpu:    &platformv1.ResourceLimits{},
							Memory: &platformv1.ResourceLimits{},
							Gpu:    &platformv1.ResourceRequest{},
						},
					},
					Extensions: map[string]*structpb.Struct{},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			InitialiseProto(tt.input)
			require.True(t, proto.Equal(tt.input, tt.expected))
		})
	}
}

func Test_NewCapsuleProto(t *testing.T) {
	c := NewCapsuleProto("project", "env", "capsule", nil)
	require.True(t, proto.Equal(c, &platformv1.Capsule{
		Kind:        CapsuleKind,
		ApiVersion:  GroupVersion.String(),
		Name:        "capsule",
		Project:     "project",
		Environment: "env",
		Spec: &platformv1.CapsuleSpec{
			Annotations: map[string]string{},
			Env: &platformv1.EnvironmentVariables{
				Raw: map[string]string{},
			},
			Scale: &platformv1.Scale{
				Horizontal: &platformv1.HorizontalScale{
					Instances: &platformv1.Instances{},
					CpuTarget: &platformv1.CPUTarget{},
				},
				Vertical: &platformv1.VerticalScale{
					Cpu:    &platformv1.ResourceLimits{},
					Memory: &platformv1.ResourceLimits{},
					Gpu:    &platformv1.ResourceRequest{},
				},
			},
			Extensions: map[string]*structpb.Struct{},
		},
	}),
	)
}

func Test_ProtoToYAML(t *testing.T) {
	c := &platformv1.CapsuleSpec{
		Annotations: map[string]string{},
		Image:       "image",
		Command:     "cmd",
		Args:        []string{},
		Env: &platformv1.EnvironmentVariables{
			Raw: map[string]string{},
			Sources: []*platformv1.EnvironmentSource{{
				Name: "name",
				Kind: "kind",
			}},
		},
		Files: []*platformv1.File{
			{
				Path:     "somepath",
				AsSecret: true,
				String_:  "hej",
				Ref:      &platformv1.FileReference{},
			},
			{
				Path: "somepath2",
				Ref: &platformv1.FileReference{
					Kind: "kiind",
					Name: "name",
					Key:  "key",
				},
			},
		},
		Interfaces: []*platformv1.CapsuleInterface{
			{
				Name:      "http",
				Port:      80,
				Liveness:  &platformv1.InterfaceLivenessProbe{},
				Readiness: &platformv1.InterfaceReadinessProbe{},
				Routes: []*platformv1.HostRoute{
					{
						Id:          "id",
						Host:        "google.com",
						Paths:       []*platformv1.HTTPPathRoute{},
						Annotations: map[string]string{},
					},
				},
			},
		},
		Extensions: map[string]*structpb.Struct{
			"value": &structpb.Struct{},
		},
		Scale: &platformv1.Scale{
			Horizontal: &platformv1.HorizontalScale{
				Instances: &platformv1.Instances{
					Min: 1,
				},
			},
			Vertical: &platformv1.VerticalScale{
				Cpu: &platformv1.ResourceLimits{},
			},
		},
	}

	s, err := ProtoToYAML(c)
	require.NoError(t, err)

	require.Equal(t, `command: cmd
env:
  sources:
  - kind: kind
    name: name
extensions:
  value: {}
files:
- asSecret: true
  path: somepath
  string: hej
- path: somepath2
  ref:
    key: key
    kind: kiind
    name: name
image: image
interfaces:
- name: http
  port: 80
  routes:
  - host: google.com
    id: id
scale:
  horizontal:
    instances:
      min: 1
`, s)
}
