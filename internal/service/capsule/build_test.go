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

	_, err := s.CreateBuild(ctx, capsuleID, "foo-bar_baz::", "", nil, nil, false)
	require.EqualError(t, err, "invalid_argument: could not parse reference: foo-bar_baz::")

	_, err = s.CreateBuild(ctx, capsuleID, "foo-bar_baz@ha256:5247f24ee94ef18029105b9a8fe2e67a021f449a7ce270ecbb451a1d42289bf6", "", nil, nil, false)
	require.EqualError(t, err, "invalid_argument: could not parse reference: foo-bar_baz@ha256:5247f24ee94ef18029105b9a8fe2e67a021f449a7ce270ecbb451a1d42289bf6")
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

	buildID, err := s.CreateBuild(ctx, capsuleID, "foobar", "", nil, nil, false)
	require.NoError(t, err)
	require.Equal(t, "index.docker.io/library/foobar:latest", buildID)

	buildID, err = s.CreateBuild(ctx, capsuleID, "foobar:hattehat", "", nil, nil, false)
	require.NoError(t, err)
	require.Equal(t, "index.docker.io/library/foobar:hattehat", buildID)
}
