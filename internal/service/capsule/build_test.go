package capsule

import (
	"context"
	"testing"

	"github.com/rigdev/rig/internal/repository"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func Test_CreateBuild_InvalidArguments(t *testing.T) {
	ctx := context.Background()
	capsuleID := uuid.New()

	s := &Service{
		logger: zaptest.NewLogger(t),
	}

	_, err := s.CreateBuild(ctx, capsuleID, "foo-bar_baz:", "", nil, nil)
	require.EqualError(t, err, "invalid_argument: invalid reference format")

	_, err = s.CreateBuild(ctx, capsuleID, "foo-bar_baz@sha256:5247f24ee94ef18029105b9a8fe2e67a021f449a7ce270ecbb451a1d42289bf6", "", nil, nil)
	require.EqualError(t, err, "invalid_argument: invalid image tag")
}

func Test_CreateBuild_ValidArguments(t *testing.T) {
	ctx := auth.WithProjectID(context.Background(), uuid.New())
	capsuleID := uuid.New()

	cr := repository.NewMockCapsule(t)
	cr.EXPECT().Get(mock.Anything, mock.Anything).Return(nil, nil)
	cr.EXPECT().CreateBuild(mock.Anything, mock.Anything, mock.Anything).Return(nil)

	s := &Service{
		cr:     cr,
		logger: zaptest.NewLogger(t),
	}

	buildID, err := s.CreateBuild(ctx, capsuleID, "foobar", "", nil, nil)
	require.NoError(t, err)
	require.Equal(t, "docker.io/library/foobar:latest", buildID)

	buildID, err = s.CreateBuild(ctx, capsuleID, "foobar:hattehat", "", nil, nil)
	require.NoError(t, err)
	require.Equal(t, "docker.io/library/foobar:hattehat", buildID)
}
