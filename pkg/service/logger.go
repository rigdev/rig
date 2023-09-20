package service

import (
	"context"
	"io"
	"net/http"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig/internal/config"
	"github.com/rigdev/rig/pkg/middleware"
	"go.uber.org/zap"
	k8s_zap "sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func NewLogger(cfg config.Config) (*zap.Logger, error) {
	return k8s_zap.NewRaw(k8s_zap.UseDevMode(true)), nil
}

type loggingMiddleware struct {
	logger *zap.Logger
}

func (i *loggingMiddleware) Wrap(next middleware.MiddlewareHandlerFunc) middleware.MiddlewareHandlerFunc {
	return func(r *http.Request) error {
		i.logger.Debug("incoming request", zap.String("path", r.URL.Path))
		err := next(r)
		if err != nil {
			i.logger.Info("incoming request error", zap.String("path", r.URL.Path), zap.Error(err))
		} else {
			i.logger.Debug("incoming request done", zap.String("path", r.URL.Path))
		}

		return err
	}
}

func (i *loggingMiddleware) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		i.logger.Debug("incoming unary request", zap.String("procedure", req.Spec().Procedure))
		res, err := next(ctx, req)
		if err != nil {
			i.logger.Info("incoming unary request error", zap.String("procedure", req.Spec().Procedure), zap.Error(err))
		} else {
			i.logger.Debug("incoming unary request done", zap.String("procedure", req.Spec().Procedure))
		}
		return res, err
	}
}

func (w *loggingMiddleware) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, s connect.Spec) connect.StreamingClientConn {
		return next(ctx, s)
	}
}

func (i *loggingMiddleware) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, c connect.StreamingHandlerConn) error {
		i.logger.Debug("incoming stream request", zap.String("procedure", c.Spec().Procedure))
		err := next(ctx, c)
		if err != nil && err != io.EOF {
			i.logger.Info("incoming stream request error", zap.String("procedure", c.Spec().Procedure), zap.Error(err))
		} else {
			i.logger.Debug("incoming stream request done", zap.String("procedure", c.Spec().Procedure))
		}
		return err
	}
}
