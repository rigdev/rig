package authentication

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/authentication"
	"github.com/rigdev/rig/pkg/errors"
)

// Delete removes a user from the namespace.
func (h *Handler) Delete(ctx context.Context, req *connect.Request[authentication.DeleteRequest]) (*connect.Response[authentication.DeleteResponse], error) {
	return nil, errors.UnimplementedErrorf("unimplemented")
}
