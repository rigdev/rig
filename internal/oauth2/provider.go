package oauth2

import (
	"context"

	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/uuid"
)

type Provider interface {
	// Validate valides the code with the oauth client and returns the user
	Validate(ctx context.Context, creds *model.ProviderCredentials, code, redirectUrl string) (*auth.Oauth2Claims, error)
	// RedirectUrl generates a redirect url for the oauth2 client
	RedirectUrl(redirectUrl, rigRedirect string, projectId uuid.UUID, creds *model.ProviderCredentials) (string, error)
	// Test validates if the oauth credentials are valid
	Test(ctx context.Context, privateKey, publicKey, redirectUrl string) error
}

type Providers struct {
	Google   Provider
	Github   Provider
	Facebook Provider
}
