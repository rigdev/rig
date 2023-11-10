package k8s_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"

	alpha1v2 "github.com/rigdev/rig/pkg/api/v1alpha2"
)

func (s *K8sTestSuite) TestCapsuleOpenAPIValidation() {
	k8sClient := s.Client
	t := s.Suite.T()

	tests := []struct {
		name         string
		capsule      *alpha1v2.Capsule
		expectedErrs field.ErrorList
	}{
		{
			name: "port number should be greater than 0",
			capsule: &alpha1v2.Capsule{
				Spec: alpha1v2.CapsuleSpec{
					Image: "test",
					Interfaces: []alpha1v2.CapsuleInterface{
						{Name: "test", Port: 0},
						{Name: "test", Port: -42},
						{
							Name: "test",
							Port: 1,
							Public: &alpha1v2.CapsulePublicInterface{
								LoadBalancer: &alpha1v2.CapsuleInterfaceLoadBalancer{
									Port: 0,
								},
							},
						},
					},
				},
			},
			expectedErrs: field.ErrorList{
				field.Invalid(
					field.NewPath("spec").Child("interfaces").Index(0).Child("port"),
					0,
					"spec.interfaces[0].port in body should be greater than or equal to 1",
				),
				field.Invalid(
					field.NewPath("spec").Child("interfaces").Index(1).Child("port"),
					-42,
					"spec.interfaces[1].port in body should be greater than or equal to 1",
				),
				field.Invalid(
					field.NewPath("spec").Child("interfaces").Index(2).Child("public").Child("loadBalancer").Child("port"),
					0,
					"spec.interfaces[2].public.loadBalancer.port in body should be greater than or equal to 1",
				),
			},
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			test.capsule.ObjectMeta = metav1.ObjectMeta{
				Name:      uuid.NewString(),
				Namespace: "default",
			}

			err := k8sClient.Create(ctx, test.capsule)
			if len(test.expectedErrs) == 0 {
				assert.NoError(t, err)
				return
			}

			var sErr *apierrors.StatusError
			if assert.ErrorAs(t, err, &sErr) {
				causes := sErr.ErrStatus.Details.Causes
				if assert.Equal(t, len(test.expectedErrs), len(causes)) {
					for i, expErr := range test.expectedErrs {
						assert.Equal(t, string(expErr.Type), string(causes[i].Type))
						assert.Equal(t, expErr.ErrorBody(), causes[i].Message)
						assert.Equal(t, expErr.Field, causes[i].Field)
					}
				}
			}
		})
	}
}
