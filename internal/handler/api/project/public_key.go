package project

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/project"
	"github.com/rigdev/rig/pkg/errors"
)

// GetPublicKeys returns all internal public keys to the client.
func (h *Handler) PublicKey(ctx context.Context, req *connect.Request[project.PublicKeyRequest]) (resp *connect.Response[project.PublicKeyResponse], err error) {
	return nil, errors.UnimplementedErrorf("Public Key")
}
