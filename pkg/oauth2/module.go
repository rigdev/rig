package oauth2

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/rigdev/rig/pkg/oauth2/facebook"
	"github.com/rigdev/rig/pkg/oauth2/github"
	"github.com/rigdev/rig/pkg/oauth2/google"
	facebookOauth2 "golang.org/x/oauth2/facebook"
	githubOauth2 "golang.org/x/oauth2/github"
	googleOauth2 "golang.org/x/oauth2/google"
)

var Module = fx.Module(
	"oauth2",
	fx.Provide(
		New,
	),
)

func New(logger *zap.Logger) (*Providers, error) {
	logger.Info("setting up oidc providers...")
	ctx := context.Background()
	newGoogle, err := newGoogle(ctx, logger)
	if err != nil {
		return nil, err
	}
	newGithub, err := newGithub(ctx, logger)
	if err != nil {
		return nil, err
	}
	newFacebook, err := newFacebook(ctx, logger)
	if err != nil {
		return nil, err
	}
	return &Providers{
		Google:   newGoogle,
		Github:   newGithub,
		Facebook: newFacebook,
	}, nil
}

func newGithub(ctx context.Context, logger *zap.Logger) (Provider, error) {
	provider, err := oidc.NewProvider(ctx, "https://token.actions.githubusercontent.com")
	if err != nil {
		return nil, err
	}
	return &github.Provider{
		Provider: provider,
		Endpoint: githubOauth2.Endpoint,
		Logger:   logger,
	}, nil
}

func newGoogle(ctx context.Context, logger *zap.Logger) (Provider, error) {
	provider, err := oidc.NewProvider(ctx, "https://accounts.google.com")
	if err != nil {
		return nil, err
	}
	return &google.Provider{
		Provider:   provider,
		Endpoint:   googleOauth2.Endpoint,
		RequireSSL: true,
		Logger:     logger,
	}, nil
}

func newFacebook(ctx context.Context, logger *zap.Logger) (Provider, error) {
	provider, err := oidc.NewProvider(ctx, "https://www.facebook.com")
	if err != nil {
		return nil, err
	}
	return &facebook.Provider{
		Provider: provider,
		Endpoint: facebookOauth2.Endpoint,
		Logger:   logger,
	}, nil
}
