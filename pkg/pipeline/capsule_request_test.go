package pipeline

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	mockclient "github.com/rigdev/rig/gen/mocks/sigs.k8s.io/controller-runtime/pkg/client"
	"github.com/rigdev/rig/pkg/api/config/v1alpha1"
	"github.com/rigdev/rig/pkg/api/v1alpha2"
	"github.com/rigdev/rig/pkg/scheme"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type pipelineTestOpts struct {
	options         []CapsuleRequestOption
	existingObjects []client.Object
}

func preparePipelineTest(t *testing.T, options pipelineTestOpts) (
	context.Context, *mockclient.MockClient, *capsuleRequest,
) {
	scheme := scheme.New()
	cc := mockclient.NewMockClient(t)
	ctx := context.Background()

	p := NewCapsulePipeline(&v1alpha1.OperatorConfig{}, scheme, logr.Discard())
	c := newCapsuleRequest(p, &v1alpha2.Capsule{}, cc, options.options...)
	for _, obj := range options.existingObjects {
		key, err := c.GetKey(obj.GetObjectKind().GroupVersionKind().GroupKind(), obj.GetName())
		require.NoError(t, err)
		c.existingObjects[key] = obj
	}

	return ctx, cc, c
}

func TestOverrideUntrackedWithForceGivesAborted(t *testing.T) {
	ctx, cc, c := preparePipelineTest(t, pipelineTestOpts{
		options: []CapsuleRequestOption{WithForce()},
	})

	sa := &v1.ServiceAccount{}
	sa.SetName("test")
	require.NoError(t, c.Set(sa))

	cc.EXPECT().Create(ctx, sa, client.DryRunAll).
		Return(kerrors.NewAlreadyExists(schema.ParseGroupResource("ServiceAccount"), sa.GetName()))

	cc.EXPECT().Get(ctx, client.ObjectKey{Name: sa.GetName()}, &v1.ServiceAccount{}).
		Return(nil)

	_, err := c.Commit(ctx)
	require.EqualError(t, err, "aborted: object exists but not in capsule status")
}

func TestOverrideUntrackedWithoutForceGivesNoop(t *testing.T) {
	ctx, cc, c := preparePipelineTest(t, pipelineTestOpts{})

	sa := &v1.ServiceAccount{}
	sa.SetName("test")
	require.NoError(t, c.Set(sa))

	cc.EXPECT().Create(ctx, sa, client.DryRunAll).
		Return(kerrors.NewAlreadyExists(schema.ParseGroupResource("ServiceAccount"), sa.GetName()))

	cc.EXPECT().Get(ctx, client.ObjectKey{Name: sa.GetName()}, &v1.ServiceAccount{}).
		Return(nil)

	cs, err := c.Commit(ctx)
	require.NoError(t, err)

	gvk := corev1.SchemeGroupVersion.WithKind("ServiceAccount")
	require.Equal(t, map[ObjectKey]*Change{
		{ObjectKey: client.ObjectKeyFromObject(sa), GroupVersionKind: gvk}: {
			state: ResourceStateAlreadyExists,
		},
	}, cs)
}

func Test_ListExisting(t *testing.T) {
	_, _, c := preparePipelineTest(t, pipelineTestOpts{
		existingObjects: []client.Object{
			&corev1.Service{ObjectMeta: metav1.ObjectMeta{
				Name: "service1",
			}},
			&corev1.Service{ObjectMeta: metav1.ObjectMeta{
				Name: "service2",
			}},
			&corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{
				Name: "serviceaccount",
			}},
		},
	})
	services, err := ListExisting(c, &corev1.Service{})
	require.NoError(t, err)
	require.Equal(t, []*corev1.Service{
		{ObjectMeta: metav1.ObjectMeta{Name: "service1"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "service2"}},
	}, services)
}

func Test_ListNew(t *testing.T) {
	_, _, c := preparePipelineTest(t, pipelineTestOpts{})

	require.NoError(t, c.Set(&corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "service1",
		},
	}))
	require.NoError(t, c.Set(&corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "service2",
		},
	}))
	require.NoError(t, c.Set(&corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: "serviceaccount",
		},
	}))

	services, err := ListNew(c, &corev1.Service{})
	require.NoError(t, err)
	require.Equal(t, []*corev1.Service{
		{ObjectMeta: metav1.ObjectMeta{Name: "service1"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "service2"}},
	}, services)
}
