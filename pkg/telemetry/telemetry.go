package telemetry

import (
	"context"
	"net/http"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig/pkg/client/segment"
	"github.com/rigdev/rig/internal/config"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/middleware"
	"github.com/rigdev/rig/pkg/uuid"
	"github.com/segmentio/analytics-go/v3"
	"go.uber.org/zap"
)

type telemetryKeyType string

const _telemetryKey telemetryKeyType = "telemetry"

var _omitPaths = map[string]struct{}{
	"/api.v1.capsule.Service/ListInstances":  {},
	"/api.v1.capsule.Service/CapsuleMetrics": {},
}

type Telemetry struct {
	logger         *zap.Logger
	sc             *segment.Client
	cfg            config.Config
	installationID uuid.UUID
}

func NewTelemetry(cfg config.Config, logger *zap.Logger, sc *segment.Client) *Telemetry {
	return &Telemetry{
		logger: logger,
		sc:     sc,
		cfg:    cfg,
	}
}

func (t *Telemetry) SetInstallationID(installationID uuid.UUID) {
	t.installationID = installationID
}

type telemetryData struct {
	userID   *uuid.UUID
	email    string
	username string
	err      error
}

func SetUserID(ctx context.Context, userID uuid.UUID) {
	if d, ok := ctx.Value(_telemetryKey).(*telemetryData); ok {
		d.userID = &userID
	}
}

func SetUserEmail(ctx context.Context, userID uuid.UUID, email string) {
	if d, ok := ctx.Value(_telemetryKey).(*telemetryData); ok {
		d.userID = &userID
		d.email = email
	}
}

func SetUserUsername(ctx context.Context, userID uuid.UUID, username string) {
	if d, ok := ctx.Value(_telemetryKey).(*telemetryData); ok {
		d.userID = &userID
		d.username = username
	}
}

func (t *Telemetry) Wrap(next middleware.MiddlewareHandlerFunc) middleware.MiddlewareHandlerFunc {
	return func(r *http.Request) error {
		d := &telemetryData{}

		defer func() {
			if !t.cfg.Telemetry.Enabled {
				return
			}

			if d.userID == nil || d.userID.IsNil() {
				return
			}

			if _, ok := _omitPaths[r.URL.Path]; ok {
				return
			}

			at := analytics.Track{
				Event: "API Request",
				Properties: analytics.NewProperties().
					SetPath(r.URL.Path).
					Set("userAgent", r.UserAgent()).
					Set("installationId", t.installationID.String()).
					Set("clusterType", t.cfg.Cluster.Type),
				UserId: d.userID.String(),
			}
			t.sc.Track(at)

			if d.err != nil {
				at := analytics.Track{
					Event: "API Error",
					Properties: analytics.NewProperties().
						SetPath(r.URL.Path).
						Set("userAgent", r.UserAgent()).
						Set("installationId", t.installationID.String()).
						Set("statusCode", errors.CodeOf(d.err).String()).
						Set("message", errors.MessageOf(d.err)),
					UserId: d.userID.String(),
				}
				t.sc.Track(at)
			}

			if d.email != "" || d.username != "" {
				ai := analytics.Identify{
					Traits: analytics.NewTraits(),
				}
				if d.email != "" {
					ai.Traits.SetEmail(d.email)
				}
				if d.username != "" {
					ai.Traits.SetUsername(d.username)
				}
				ai.Traits.Set("installationId", t.installationID.String())
				ai.UserId = d.userID.String()
				t.sc.Identify(ai)
			}
		}()

		return next(r.WithContext(context.WithValue(r.Context(), _telemetryKey, d)))
	}
}

func (t *Telemetry) onError(ctx context.Context, err error) {
	if err == nil {
		return
	}

	if !t.cfg.Telemetry.Enabled {
		return
	}

	d, ok := ctx.Value(_telemetryKey).(*telemetryData)
	if !ok {
		return
	}

	d.err = err
}

func (t *Telemetry) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		res, err := next(ctx, req)
		t.onError(ctx, err)
		return res, err
	}
}

func (t *Telemetry) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, s connect.Spec) connect.StreamingClientConn {
		return next(ctx, s)
	}
}

func (t *Telemetry) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, c connect.StreamingHandlerConn) error {
		err := next(ctx, c)
		t.onError(ctx, err)
		return err
	}
}
