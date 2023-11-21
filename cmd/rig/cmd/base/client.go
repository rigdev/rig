package base

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/golang-jwt/jwt"
	"github.com/rigdev/rig-go-api/api/v1/project"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/rig/cmd/cmdconfig"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

const (
	_rigProjectTokenHeader = "X-Rig-Project-Token"
)

var _omitProjectToken = map[string]struct{}{
	"/api.v1.project.Service/Use":    {},
	"/api.v1.project.Service/Create": {},
	"/api.v1.project.Service/List":   {},

	"/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo": {},
}

var clientModule = fx.Module("client",
	fx.Supply(&http.Client{}),
	fx.Provide(func(
		ctx context.Context,
		cmd *cobra.Command,
		s *cmdconfig.Service,
		cfg *cmdconfig.Config,
	) (rig.Client, error) {

		ai := &authInterceptor{cfg: cfg}
		rigClient := rig.NewClient(
			rig.WithHost(s.Server),
			rig.WithInterceptors(ai, &userAgentInterceptor{}),
			rig.WithSessionManager(&configSessionManager{cfg: cfg}),
		)
		ai.rig = rigClient
		if err := CheckAuth(ctx, cmd, rigClient, cfg); err != nil {
			return nil, err
		}

		return rigClient, nil
	}),
	fx.Provide(func(cfg *cmdconfig.Config) []connect.Interceptor {
		return []connect.Interceptor{&userAgentInterceptor{}, &authInterceptor{cfg: cfg}}
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

type authInterceptor struct {
	cfg *cmdconfig.Config
	rig rig.Client
}

func (i *authInterceptor) handleAuth(ctx context.Context, h http.Header, method string) {
	if _, ok := _omitProjectToken[method]; !ok {
		i.setProjectToken(ctx, h)
	}
}

func (i *authInterceptor) setProjectToken(ctx context.Context, h http.Header) {
	if i.cfg.GetCurrentContext().Project.ProjectToken == "" {
		return
	}

	c := jwt.StandardClaims{}
	p := jwt.Parser{
		SkipClaimsValidation: true,
	}
	_, _, err := p.ParseUnverified(
		i.cfg.GetCurrentContext().Project.ProjectToken,
		&c,
	)
	if err != nil {
		return
	}

	// Don't use if invalid user id.
	if i.cfg.GetCurrentAuth().UserID.String() != c.Subject {
		return
	}

	if !c.VerifyExpiresAt(time.Now().Add(30*time.Second).Unix(), true) &&
		i.cfg.GetCurrentContext().Project.ProjectID != "" {
		res, err := i.rig.Project().Use(ctx, &connect.Request[project.UseRequest]{
			Msg: &project.UseRequest{
				ProjectId: i.cfg.GetCurrentContext().Project.ProjectID,
			},
		})
		if err == nil {
			i.cfg.GetCurrentContext().Project.ProjectToken = res.Msg.GetProjectToken()
			if err := i.cfg.Save(); err != nil {
				fmt.Fprintf(os.Stderr, "error saving config: %v\n", err)
			}
		}
	}

	h.Set(_rigProjectTokenHeader, fmt.Sprint(i.cfg.GetCurrentContext().Project.ProjectToken))
}

func (i *authInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, ar connect.AnyRequest) (connect.AnyResponse, error) {
		i.handleAuth(ctx, ar.Header(), ar.Spec().Procedure)
		return next(ctx, ar)
	}
}

func (i *authInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, s connect.Spec) connect.StreamingClientConn {
		conn := next(ctx, s)
		i.handleAuth(ctx, conn.RequestHeader(), s.Procedure)
		return conn
	}
}

func (i *authInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, shc connect.StreamingHandlerConn) error {
		i.handleAuth(ctx, shc.RequestHeader(), shc.Spec().Procedure)
		return next(ctx, shc)
	}
}
