package roclient

import (
	"context"
	"fmt"
	"testing"

	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/obj"
	"github.com/rigdev/rig/pkg/scheme"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type getCase struct {
	name        string
	key         client.ObjectKey
	emptyObj    client.Object
	expected    client.Object
	expectedErr error
	opts        []client.GetOption
}

type listCase struct {
	name        string
	emptyList   client.ObjectList
	expected    client.ObjectList
	expectedErr error
	opts        []client.ListOption
	kind        string
}

func TestReader(t *testing.T) {
	reader1 := NewReader(scheme.New())
	reader2 := NewReader(scheme.New())
	reader3 := NewReader(scheme.New())

	deployment1 := &appsv1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name:      "deployment1",
			Namespace: "lolcat",
		},
	}
	deployment1.SetGroupVersionKind(appsv1.SchemeGroupVersion.WithKind("Deployment"))

	deployment2 := &appsv1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name:      "deployment2",
			Namespace: "lolcat",
		},
	}
	deployment2.SetGroupVersionKind(appsv1.SchemeGroupVersion.WithKind("Deployment"))

	configMap1 := &corev1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name:      "configmap1",
			Namespace: "lolcat",
		},
	}
	configMap1.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("ConfigMap"))

	configMap2 := &corev1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name:      "configmap2",
			Namespace: "lolcat",
		},
	}
	configMap2.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("ConfigMap"))

	configMap3 := &corev1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name:      "configmap3",
			Namespace: "lolcat",
		},
	}
	configMap3.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("ConfigMap"))

	service1 := &corev1.Service{
		ObjectMeta: v1.ObjectMeta{
			Name:      "service1",
			Namespace: "lolcat",
		},
	}
	service1.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("Service"))

	service3 := service1.DeepCopy()

	require.NoError(t, reader1.AddObject(deployment1))
	require.NoError(t, reader2.AddObject(deployment2))
	require.NoError(t, reader1.AddObject(configMap1))
	require.NoError(t, reader2.AddObject(configMap2))
	require.NoError(t, reader3.AddObject(configMap3))
	require.NoError(t, reader1.AddObject(service1))
	require.NoError(t, reader3.AddObject(service3))

	r := NewLayeredReader(reader1, reader2, reader3)

	getCases := []getCase{
		{
			name: "Get Deployment",
			key: client.ObjectKey{
				Name:      "deployment1",
				Namespace: "lolcat",
			},
			emptyObj:    &appsv1.Deployment{},
			expected:    deployment1,
			expectedErr: nil,
		},
		{
			name: "Not Found",
			key: client.ObjectKey{
				Name:      "deployment3",
				Namespace: "lolcat",
			},
			emptyObj: &appsv1.Deployment{},
			expected: &appsv1.Deployment{},
			expectedErr: kerrors.NewNotFound(
				schema.GroupResource{Group: (&appsv1.Deployment{}).GetObjectKind().GroupVersionKind().Group},
				"deployment3"),
		},
	}

	listCases := []listCase{
		{
			name: "List Deployment",
			emptyList: &appsv1.DeploymentList{
				TypeMeta: v1.TypeMeta{
					Kind:       "DeploymentList",
					APIVersion: "apps/v1",
				},
			},
			expected: &appsv1.DeploymentList{
				Items: []appsv1.Deployment{*deployment1, *deployment2},
			},
			expectedErr: nil,
			kind:        "Deployment",
		},
		{
			name:      "List ConfigMaps",
			emptyList: &corev1.ConfigMapList{},
			expected: &corev1.ConfigMapList{
				Items: []corev1.ConfigMap{*configMap1, *configMap2, *configMap3},
			},
			expectedErr: nil,
			kind:        "ConfigMap",
		},
		{
			name:      "List Services",
			emptyList: &corev1.ServiceList{},
			expected: &corev1.ServiceList{
				Items: []corev1.Service{*service1},
			},
			expectedErr: nil,
			kind:        "Service",
		},
	}

	for _, c := range getCases {
		t.Run(c.name, func(t *testing.T) {
			err := r.Get(context.Background(), c.key, c.emptyObj, c.opts...)
			if errors.CodeOf(err) != errors.CodeOf(c.expectedErr) || errors.MessageOf(err) != errors.MessageOf(c.expectedErr) {
				t.Errorf("expected %v, got %v", c.expectedErr, err)
			}

			c := obj.NewComparison(c.emptyObj, c.expected, scheme.New())
			d, err := c.ComputeDiff()
			require.NoError(t, err)
			require.Equal(t, 0, len(d.Report.Diffs))
		})
	}

	for _, c := range listCases {
		t.Run(c.name, func(t *testing.T) {
			err := r.List(context.Background(), c.emptyList, c.opts...)
			if errors.CodeOf(err) != errors.CodeOf(c.expectedErr) || errors.MessageOf(err) != errors.MessageOf(c.expectedErr) {
				t.Errorf("expected %v, got %v", c.expectedErr, err)
			}

			compareLists(t, c.emptyList, c.expected)
			require.NoError(t, err)
		})
	}
}

func compareLists(t *testing.T, got, expected client.ObjectList) {
	switch v := got.(type) {
	case *appsv1.DeploymentList:
		require.Equal(t, len(expected.(*appsv1.DeploymentList).Items), len(v.Items))
	case *corev1.ConfigMapList:
		require.Equal(t, len(expected.(*corev1.ConfigMapList).Items), len(v.Items))
	case *corev1.ServiceList:
		require.Equal(t, len(expected.(*corev1.ServiceList).Items), len(v.Items))
	default:
		require.Fail(t, fmt.Sprintf("unexpected type %T", v))
	}
}
