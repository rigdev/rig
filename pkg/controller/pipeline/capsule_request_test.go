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
	v1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func preparePipelineTest(t *testing.T, opts ...CapsuleRequestOption) (
	context.Context, *mockclient.MockClient, *capsuleRequest,
) {
	scheme := scheme.New()
	cc := mockclient.NewMockClient(t)
	ctx := context.Background()

	p := New(cc, &v1alpha1.OperatorConfig{}, scheme, logr.Discard())
	c := newCapsuleRequest(p, &v1alpha2.Capsule{}, opts...)

	return ctx, cc, c
}

func TestOverrideUntrackedWithForceGivesAborted(t *testing.T) {
	ctx, cc, c := preparePipelineTest(t, WithForce())

	sa := &v1.ServiceAccount{}
	sa.SetName("test")
	require.NoError(t, c.Set(sa))

	cc.EXPECT().Create(ctx, sa, client.DryRunAll).
		Return(kerrors.NewAlreadyExists(schema.ParseGroupResource("ServiceAccount"), sa.GetName()))

	cc.EXPECT().Get(ctx, client.ObjectKey{Name: sa.GetName()}, &v1.ServiceAccount{}).
		Return(nil)

	_, err := c.commit(ctx)
	require.EqualError(t, err, "aborted: object exists but not in capsule status")
}

func TestOverrideUntrackedWithoutForceGivesNoop(t *testing.T) {
	ctx, cc, c := preparePipelineTest(t)

	sa := &v1.ServiceAccount{}
	sa.SetName("test")
	require.NoError(t, c.Set(sa))

	cc.EXPECT().Create(ctx, sa, client.DryRunAll).
		Return(kerrors.NewAlreadyExists(schema.ParseGroupResource("ServiceAccount"), sa.GetName()))

	cc.EXPECT().Get(ctx, client.ObjectKey{Name: sa.GetName()}, &v1.ServiceAccount{}).
		Return(nil)

	cs, err := c.commit(ctx)
	require.NoError(t, err)

	require.Equal(t, map[objectKey]*change{
		{ObjectKey: client.ObjectKeyFromObject(sa), GroupVersionKind: CoreServiceAccount}: {
			state: ResourceStateAlreadyExists,
		},
	}, cs)
}
