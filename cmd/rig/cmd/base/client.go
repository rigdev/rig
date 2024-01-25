package base

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/rig/cmd/cmdconfig"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/rigdev/rig/cmd/rig/services/auth"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var clientModule = fx.Module("client",
	fx.Supply(&http.Client{}),
	fx.Provide(newRigClient),
	fx.Provide(func(cfg *cmdconfig.Config) []connect.Interceptor {
		return []connect.Interceptor{&userAgentInterceptor{}}
	}),
)

func newRigClient(
	cmd *cobra.Command,
	s *cmdconfig.Service,
	cfg *cmdconfig.Config,
	interactive Interactive,
) (rig.Client, *auth.Service, error) {
	options := []rig.Option{
		rig.WithInterceptors(&userAgentInterceptor{}),
		rig.WithSessionManager(&configSessionManager{cfg: cfg}),
	}

	if flags.Flags.BasicAuth {
		options = append(options, rig.WithBasicAuthOption(rig.ClientCredential{}))
	}

	if flags.Flags.Host != "" {
		options = append(options, rig.WithHost(flags.Flags.Host))
	} else {
		options = append(options, rig.WithHost(s.Server))
	}

	r := rig.NewClient(options...)
	a := auth.NewService(r, cfg)

	if !SkipChecks(cmd) {
		if err := a.CheckAuth(context.TODO(), cmd, bool(interactive)); err != nil {
			return nil, nil, err
		}
	}

	return r, a, nil
}

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
