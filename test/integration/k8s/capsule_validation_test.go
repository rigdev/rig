package k8s_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	alpha1v1 "github.com/rigdev/rig/pkg/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func TestIntegrationCapsuleOpenAPIValidation(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	env := setupTest(t, options{})
	defer env.cancel()
	k8sClient := env.k8sClient

	tests := []struct {
		name         string
		capsule      *alpha1v1.Capsule
		expectedErrs field.ErrorList
	}{
		{
			name: "port number should be greater than 0",
			capsule: &alpha1v1.Capsule{
				Spec: alpha1v1.CapsuleSpec{
					Image: "test",
					Interfaces: []alpha1v1.CapsuleInterface{
						{Name: "test", Port: 0},
						{Name: "test", Port: -42},
						{
							Name: "test",
							Port: 1,
							Public: &alpha1v1.CapsulePublicInterface{
								LoadBalancer: &alpha1v1.CapsuleInterfaceLoadBalancer{
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
