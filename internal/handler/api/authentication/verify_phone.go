package authentication

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/authentication"
	"github.com/rigdev/rig/pkg/errors"
)

func (h *Handler) VerifyPhoneNumber(ctx context.Context, req *connect.Request[authentication.VerifyPhoneNumberRequest]) (*connect.Response[authentication.VerifyPhoneNumberResponse], error) {
	return nil, errors.UnimplementedErrorf("unimplemented")
}
