package cli

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/rig/cmd/cmdconfig"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/rigdev/rig/pkg/cli/scope"
)

func GetClientOptions(cfg *cmdconfig.Config) ([]rig.Option, error) {
	options := []rig.Option{
		rig.WithInterceptors(&userAgentInterceptor{}),
		rig.WithSessionManager(&configSessionManager{cfg: cfg}),
	}

	if flags.Flags.BasicAuth {
		options = append(options, rig.WithBasicAuthOption(rig.ClientCredential{}))
	}

	host := flags.Flags.Host
	// fmt.Printf("getClientOptions %s, %+v\n", host, cfg.GetCurrentCotext())
	if host == "" {
		if rCtx := cfg.GetCurrentContext(); rCtx != nil {
			if svc := rCtx.GetService(); svc != nil {
				host = svc.Server
			}
		}
	}

	options = append(options, rig.WithHost(host))
	return options, nil
}

func newRigClient(scope scope.Scope) (rig.Client, error) {
	opts, err := GetClientOptions(scope.GetCfg())
	if err != nil {
		return nil, err
	}

	return rig.NewClient(opts...), nil
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
	return s.cfg.GetCurrentContext().GetAuth().AccessToken
}

func (s *configSessionManager) GetRefreshToken() string {
	return s.cfg.GetCurrentContext().GetAuth().RefreshToken
}

func (s *configSessionManager) SetAccessToken(accessToken, refreshToken string) {
	s.cfg.GetCurrentContext().GetAuth().AccessToken = accessToken
	s.cfg.GetCurrentContext().GetAuth().RefreshToken = refreshToken
	if err := s.cfg.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "error saving config: %v\n", err)
	}
}
