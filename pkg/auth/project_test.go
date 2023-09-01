package auth

import (
	"context"
	"testing"

	"github.com/rigdev/rig/pkg/uuid"
	"github.com/stretchr/testify/require"
)

func Test_ProjectID(t *testing.T) {
	ctx := context.TODO()

	pID1 := uuid.New()
	ctx = WithProjectID(ctx, pID1)

	pID2, err := GetProjectID(ctx)
	require.Equal(t, pID1, pID2)
	require.NoError(t, err)
}

func Test_ProjectID_Missing(t *testing.T) {
	ctx := context.TODO()

	pID1, err := GetProjectID(ctx)
	require.Equal(t, uuid.Nil, pID1)
	require.EqualError(t, err, "permission_denied: no project selected")
}
