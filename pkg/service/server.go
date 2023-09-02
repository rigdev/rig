package service

import (
	"context"
	"embed"
	"encoding/json"
	goerrors "errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/bufbuild/connect-go"
	grpcreflect "github.com/bufbuild/connect-grpcreflect-go"
	"github.com/go-chi/chi/v5"
	"github.com/rigdev/rig/internal/config"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/middleware"
	"github.com/rigdev/rig/pkg/telemetry"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

type HTTPHandler interface {
	Build() (string, string, HandlerFunc)
}

type GRPCHandler interface {
	ServiceName() string
	Build(opts ...connect.HandlerOption) (string, http.Handler)
}

type Server struct {
	logger *zap.Logger
	cfg    config.Config
	addr   string
	r      *chi.Mux
	srv    *http.Server
	a      *Authorization
	t      *telemetry.Telemetry
	mw     []middleware.Middleware
	p      NewServerParams
}

type NewServerParams struct {
	fx.In

	Lifecycle      fx.Lifecycle
	Config         config.Config
	Logger         *zap.Logger
	Authentication *Authorization
	Telemetry      *telemetry.Telemetry

	GRPCHandlers []GRPCHandler `group:"grpc_handlers"`
	HTTPHandlers []HTTPHandler `group:"http_handlers"`
}

func NewServer(p NewServerParams) *Server {
	s := &Server{
		logger: p.Logger,
		cfg:    p.Config,
		srv: &http.Server{
			ReadHeaderTimeout: time.Second,
			ReadTimeout:       5 * time.Minute,
			WriteTimeout:      5 * time.Minute,
			MaxHeaderBytes:    8 * 1024, // 8KiB
		},
		r: chi.NewRouter(),
		a: p.Authentication,
		t: p.Telemetry,
		p: p,
	}

	p.Lifecycle.Append(fx.StartStopHook(s.Start, s.Stop))

	return s
}

func (s *Server) Init() {
	s.mw = append(s.mw, s.t)
	s.mw = append(s.mw, &loggingMiddleware{
		logger: s.logger,
	})
	s.mw = append(s.mw, s.a)

	var ns []string
	for _, h := range s.p.GRPCHandlers {
		ns = append(ns, h.ServiceName())
		s.AddGRPCHandler(h.Build(s.Interceptors()))
	}
	reflector := grpcreflect.NewStaticReflector(ns...)
	s.AddGRPCHandler(grpcreflect.NewHandlerV1(reflector))
	s.AddGRPCHandler(grpcreflect.NewHandlerV1Alpha(reflector))

	for _, h := range s.p.HTTPHandlers {
		s.AddHTTPHandler(h.Build())
	}
}

func (s *Server) AddGRPCHandler(p string, h http.Handler) {
	if strings.HasSuffix(p, "/") {
		p = path.Join(p, "*")
	}

	s.AddHTTPHandler(http.MethodPost, p, func(w http.ResponseWriter, r *http.Request) error {
		h.ServeHTTP(w, r)
		return nil
	})
}

func (s *Server) AddHTTPHandler(method, p string, h HandlerFunc) {
	s.logger.Info("adding http handler", zap.String("method", method), zap.String("path", p))
	f := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hs := middleware.MiddlewareHandlerFunc(func(r *http.Request) error {
			return h(w, r)
		})

		for i := len(s.mw) - 1; i >= 0; i-- {
			hs = s.mw[i].Wrap(hs)
		}

		if err := hs(r); err != nil {
			handleError(w, r, err)
		}
	})

	s.r.Method(method, p, f)
}

func handleError(w http.ResponseWriter, r *http.Request, err error) {
	if ew := connect.NewErrorWriter(); ew.IsSupported(r) {
		ew.Write(w, r, err)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(errors.ToHTTP(err))
		body := struct {
			Code    connect.Code `json:"code"`
			Message string       `json:"message"`
		}{
			Code:    errors.CodeOf(err),
			Message: errors.MessageOf(err),
		}
		bs, _ := json.Marshal(body)
		w.Write(bs)
		w.Write([]byte{'\n'})
	}
}

//go:embed all:web
var embeddedFS embed.FS

func (s *Server) EmbeddedFileServer() {
	s.logger.Info("serving web from embedded files")

	subFS, err := fs.Sub(embeddedFS, "web")
	if err != nil {
		s.logger.Fatal("could not get sub fs from embedded fs", zap.Error(err))
	}

	h := http.FileServer(http.FS(subFS))
	s.r.Method(http.MethodGet, "/_nuxt/*", h)
	s.r.Method(http.MethodGet, "/favicon.ico", h)
	s.r.Method(http.MethodGet, "/200.html", h)
	s.r.Method(http.MethodGet, "/404.html", h)

	idxF, err := embeddedFS.Open("web/index.html")
	if err != nil {
		s.logger.Fatal("could not open index.html from embedded fs", zap.Error(err))
	}
	defer idxF.Close()

	bs, err := io.ReadAll(idxF)
	if err != nil {
		s.logger.Fatal("could not read index.html from embedded fs", zap.Error(err))
	}

	s.r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		w.Write(bs)
	})
}

func (s *Server) Interceptors() connect.Option {
	return connect.WithInterceptors(
		s.t,
		&loggingMiddleware{
			logger: s.logger,
		},
	)
}

func (s *Server) Start() error {
	s.srv.Handler = h2c.NewHandler(
		s.r,
		&http2.Server{},
	)

	l, err := net.Listen("tcp", fmt.Sprint(":", s.cfg.Port))
	if err != nil {
		return err
	}

	s.addr = l.Addr().String()
	s.logger.Info("running server", zap.String("addr", s.addr))

	go func() {
		if err := s.srv.Serve(l); err != nil && !goerrors.Is(err, http.ErrServerClosed) {
			s.logger.Fatal("error running server", zap.Error(err))
		}
	}()

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("stopping server", zap.String("addr", s.addr))
	s.srv.Shutdown(ctx)
	return nil
}

func (s *Server) Address() string {
	return s.addr
}
