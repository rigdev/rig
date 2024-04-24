package errors

import (
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Test_Grpc_Errors(t *testing.T) {
	e := NotFoundErrorf("test")
	require.True(t, IsNotFound(e))
	require.Equal(t, codes.NotFound, status.Code(e))
	require.Equal(t, connect.CodeNotFound, connect.CodeOf(e))
	require.Equal(t, "test", MessageOf(e))
	require.Equal(t, "not_found: test", e.Error())

	connectErr := connect.NewError(connect.CodeNotFound, e)
	require.Equal(t, codes.NotFound, status.Code(connectErr))
	require.True(t, IsNotFound(connectErr))
	require.Equal(t, connect.CodeNotFound, connect.CodeOf(connectErr))
	require.Equal(t, "not_found: test", MessageOf(connectErr))
	require.Equal(t, "not_found: not_found: test", connectErr.Error())

	grpcError := status.Error(codes.NotFound, e.Error())
	require.Equal(t, codes.NotFound, status.Code(grpcError))
	require.True(t, IsNotFound(grpcError))
	require.Equal(t, "not_found: test", MessageOf(grpcError))
	require.Equal(t, "rpc error: code = NotFound desc = not_found: test", grpcError.Error())

}
