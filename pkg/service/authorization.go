package service

import (
	"context"
	"net/http"
	"strings"

	service_auth "github.com/rigdev/rig/internal/service/auth"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/middleware"
	"github.com/rigdev/rig/pkg/telemetry"
	"go.uber.org/zap"
)

var OmitAuth = map[string]struct{}{
	"/oauth/callback": {},

	"/api.v1.authentication.Service/Login":             {},
	"/api.v1.authentication.Service/Register":          {},
	"/api.v1.authentication.Service/VerifyEmail":       {},
	"/api.v1.authentication.Service/RefreshToken":      {},
	"/api.v1.authentication.Service/OauthCallback":     {},
	"/api.v1.authentication.Service/SendPasswordReset": {},
	"/api.v1.authentication.Service/ResetPassword":     {},
	"/api.v1.authentication.Service/GetAuthConfig":     {},

	"/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo": {},

	"/api/v1/status": {},
}

var OmitProjectToken = map[string]struct{}{
	"/api.v1.project.Service/Use":    {},
	"/api.v1.project.Service/Create": {},
	"/api.v1.project.Service/List":   {},

	"/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo": {},
}

var AllUsersAllow = map[string]struct{}{
	"/api.v1.authentication.Service": {},
}

const (
	RigProjectTokenHeader = "X-Rig-Project-Token"
)

type Authorization struct {
	logger *zap.Logger
	as     *service_auth.Service
}

func NewAuthorization(logger *zap.Logger, as *service_auth.Service) *Authorization {
	return &Authorization{
		logger: logger,
		as:     as,
	}
}

func (a *Authorization) Wrap(next middleware.MiddlewareHandlerFunc) middleware.MiddlewareHandlerFunc {
	return func(r *http.Request) error {
		ctx, err := a.handleAuth(r.Context(), r.URL.Path, r.Header)
		if err != nil {
			return err
		}
		return next(r.WithContext(ctx))
	}
}

func (a *Authorization) handleAuth(ctx context.Context, path string, h http.Header) (context.Context, error) {
	logger := a.logger.With(zap.String("path", path))
	if _, ok := OmitAuth[path]; ok {
		logger.Debug("skipping auth check for request")
		return ctx, nil
	}

	jwt := h.Get("Authorization")
	if !strings.HasPrefix(jwt, "Bearer ") {
		logger.Debug("request is missing authorization bearer")
		return ctx, errors.UnauthenticatedErrorf("missing authorization bearer")
	}

	jwt = strings.TrimPrefix(jwt, "Bearer ")

	c, err := a.as.ValidateAccessToken(ctx, jwt)
	if err != nil {
		logger.Debug("invalid auth token", zap.Error(err))
		return ctx, errors.UnauthenticatedErrorf("%v", err)
	}

	if c.GetProjectID() == "" || c.GetSubject().IsNil() || c.GetSubjectType() == auth.SubjectTypeInvalid {
		logger.Debug("auth token missing essential properties")
		return nil, errors.UnauthenticatedErrorf("invalid auth token content")
	}

	switch c.GetSubjectType() {
	case auth.SubjectTypeServiceAccount:
		// Credentials are constrained to project they are created for.
		ctx = auth.WithProjectID(ctx, c.ProjectID)

	case auth.SubjectTypeUser:
		switch c.GetProjectID() {
		case auth.RigProjectID:
			// Rig user, set project ID from project token.
			telemetry.SetUserID(ctx, c.Subject)
			pToken := h.Get(RigProjectTokenHeader)
			if pToken != "" {
				pc, err := a.as.ValidateProjectToken(ctx, pToken)
				if err != nil {
					a.logger.Debug("invalid project token", zap.Error(err))
					return nil, err
				}

				logger = logger.With(zap.String("project_id", pc.UseProjectID))

				if c.GetProjectID() != pc.ProjectID || c.GetSubject() != pc.GetSubject() {
					logger.Debug("project token claims doesn't match user access token claims")
					return nil, errors.UnauthenticatedErrorf("invalid project token content")
				}

				ctx = auth.WithProjectID(ctx, pc.UseProjectID)
			} else {
				logger.Debug("request is missing 'X-Rig-Project-Token' token")
			}

		default:
			// Users only have access to the limited auth API set.
			if err := checkUserAccess(path, logger); err != nil {
				return nil, err
			}

			// Project user, set project ID to jwt project.
			ctx = auth.WithProjectID(ctx, c.ProjectID)
		}

	default:
		logger.Debug("invalid subject type", zap.Int("subject_type", int(c.GetSubjectType())))
		return nil, errors.UnauthenticatedErrorf("invalid auth token content")
	}

	ctx = auth.WithClaims(ctx, c)

	return ctx, nil
}

func checkUserAccess(path string, logger *zap.Logger) error {
	for p := range AllUsersAllow {
		if strings.HasPrefix(path, p) {
			return nil
		}
	}

	logger.Debug("project user attempted to access Rig management SDKs")
	return errors.PermissionDeniedErrorf("access denied")
}
