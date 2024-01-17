package base

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/rig/cmd/cmdconfig"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var clientModule = fx.Module("client",
	fx.Supply(&http.Client{}),
	fx.Provide(func(
		ctx context.Context,
		cmd *cobra.Command,
		s *cmdconfig.Service,
		cfg *cmdconfig.Config,
		interactive bool,
	) (rig.Client, error) {
		rigClient := rig.NewClient(
			rig.WithHost(s.Server),
			rig.WithInterceptors(&userAgentInterceptor{}),
			rig.WithSessionManager(&configSessionManager{cfg: cfg}),
		)

		if err := CheckAuth(ctx, cmd, rigClient, cfg, interactive); err != nil {
			return nil, err
		}

		return rigClient, nil
	}),
	fx.Provide(func(cfg *cmdconfig.Config) []connect.Interceptor {
		return []connect.Interceptor{&userAgentInterceptor{}}
	}),
)

type userAgentInterceptor struct{}

func (i *userAgentInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, ar connect.AnyRequest) (connect.AnyResponse, error) {
		i.setUserAgent(ar.Header())
		return next(ctx, ar)
	}
}

func (i *userAgentInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, s connect.Spec) connect.StreamingClientConn {
		conn := next(ctx, s)
		i.setUserAgent(conn.RequestHeader())
		return conn
	}
}

func (i *userAgentInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, shc connect.StreamingHandlerConn) error {
		i.setUserAgent(shc.RequestHeader())
		return next(ctx, shc)
	}
}

func (i *userAgentInterceptor) setUserAgent(h http.Header) {
	h.Set("User-Agent", "Rig-CLI/v0.0.1")
}

type configSessionManager struct {
	cfg *cmdconfig.Config
}

func (s *configSessionManager) GetAccessToken() string {
	return s.cfg.GetCurrentAuth().AccessToken
}

func (s *configSessionManager) GetRefreshToken() string {
	return s.cfg.GetCurrentAuth().RefreshToken
}

func (s *configSessionManager) SetAccessToken(accessToken, refreshToken string) {
	s.cfg.GetCurrentAuth().AccessToken = accessToken
	s.cfg.GetCurrentAuth().RefreshToken = refreshToken
	if err := s.cfg.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "error saving config: %v\n", err)
	}
}
