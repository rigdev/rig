package cli

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/rigdev/rig/cmd/rig/services/auth"
	"github.com/rigdev/rig/pkg/cli/scope"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var clientModule = fx.Module("client",
	fx.Supply(&http.Client{}),
	fx.Provide(newRigClient),
	fx.Provide(func() []connect.Interceptor {
		return []connect.Interceptor{&userAgentInterceptor{}}
	}),
)

func newRigClient(
	cmd *cobra.Command,
	scope scope.Scope,
) (rig.Client, *auth.Service, error) {
	options := []rig.Option{
		rig.WithInterceptors(&userAgentInterceptor{}),
		rig.WithSessionManager(&configSessionManager{scope: scope}),
	}

	if flags.Flags.BasicAuth {
		options = append(options, rig.WithBasicAuthOption(rig.ClientCredential{}))
	}

	if flags.Flags.Host != "" {
		options = append(options, rig.WithHost(flags.Flags.Host))
	} else {
		options = append(options, rig.WithHost(scope.GetCurrentContext().GetService().Server))
	}

	r := rig.NewClient(options...)
	a := auth.NewService(r, scope)

	if !SkipFX(cmd) {
		if err := a.CheckAuth(context.TODO(), cmd, scope.IsInteractive(), flags.Flags.BasicAuth); err != nil {
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
	scope scope.Scope
}

func (s *configSessionManager) GetAccessToken() string {
	return s.scope.GetCurrentContext().GetAuth().AccessToken
}

func (s *configSessionManager) GetRefreshToken() string {
	return s.scope.GetCurrentContext().GetAuth().RefreshToken
}

func (s *configSessionManager) SetAccessToken(accessToken, refreshToken string) {
	s.scope.GetCurrentContext().GetAuth().AccessToken = accessToken
	s.scope.GetCurrentContext().GetAuth().RefreshToken = refreshToken
	if err := s.scope.GetCfg().Save(); err != nil {
		fmt.Fprintf(os.Stderr, "error saving config: %v\n", err)
	}
}
