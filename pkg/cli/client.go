package cli

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/rig/cmd/cmdconfig"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/rigdev/rig/pkg/errors"
)

func getClientOptions(cfg *cmdconfig.Config) ([]rig.Option, error) {
	options := []rig.Option{
		rig.WithInterceptors(&userAgentInterceptor{}),
		rig.WithSessionManager(&configSessionManager{cfg: cfg}),
	}

	if flags.Flags.BasicAuth {
		options = append(options, rig.WithBasicAuthOption(rig.ClientCredential{}))
	}

	host := flags.Flags.Host

	if host == "" {
		host = cfg.GetCurrentContext().GetService().Server
	}
	url, err := url.Parse(host)
	if err != nil {
		return nil, errors.InvalidArgumentErrorf("invalid host, must be a fully qualified URL: %v", err)
	}

	if url.Scheme != "http" && url.Scheme != "https" {
		return nil, errors.InvalidArgumentErrorf("invalid host, must start with `https://` or `http://`")
	}

	if url.Host == "" {
		return nil, errors.InvalidArgumentErrorf("invalid host, must be a fully qualified URL: missing hostname")
	}

	options = append(options, rig.WithHost(host))
	return options, nil
}

func newRigClient(cfg *cmdconfig.Config) (rig.Client, error) {
	opts, err := getClientOptions(cfg)
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
