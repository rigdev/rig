package authentication

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/authentication"
	"github.com/rigdev/rig/pkg/auth"
	"k8s.io/utils/strings/slices"
)

// GetAuthConfig this returns a config for a specific namespace together with urls to all OIDC providers.
func (h *Handler) GetAuthConfig(ctx context.Context, req *connect.Request[authentication.GetAuthConfigRequest]) (resp *connect.Response[authentication.GetAuthConfigResponse], err error) {
	pID := req.Msg.GetProjectId()
	ctx = auth.WithProjectID(ctx, pID)

	p, err := h.ps.GetProject(ctx)
	if err != nil {
		return nil, err
	}

	us, err := h.us.GetSettings(ctx)
	if err != nil {
		return nil, err
	}

	resp = &connect.Response[authentication.GetAuthConfigResponse]{
		Msg: &authentication.GetAuthConfigResponse{
			Name:             p.GetName(),
			ValidatePassword: true,
			LoginTypes:       us.GetLoginMechanisms(),
			AllowsRegister:   us.GetAllowRegister(),
		},
	}

	if !slices.Contains(us.GetOauthSettings().GetCallbackUrls(), req.Msg.GetRedirectAddr()) {
		return resp, nil
	}

	oauthProviders, err := h.as.GetOauth2Providers(ctx, req.Msg.GetRedirectAddr())
	if err != nil {
		fmt.Println(err)
		return resp, nil
	}
	resp.Msg.OauthProviders = oauthProviders

	return resp, nil
}
