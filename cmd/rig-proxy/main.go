package main

import (
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	proto_proxy "github.com/rigdev/rig/gen/go/proxy"
	"github.com/rigdev/rig/internal/build"
	"github.com/rigdev/rig/internal/service/auth"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/proxy"
	"github.com/rigdev/rig/pkg/uuid"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/protobuf/encoding/protojson"
	"inet.af/tcpproxy"
)

func main() {
	var printVersion bool
	flag.BoolVar(&printVersion, "version", false, "show version")
	flag.Parse()
	if printVersion {
		fmt.Print(build.VersionStringFull())
		return
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}

	pc := &proto_proxy.Config{}
	if env, ok := os.LookupEnv("RIG_PROXY_CONFIG"); ok {
		env, err := strconv.Unquote(env)
		if err != nil {
			logger.Fatal("invalid format of RIG_PROXY_CONFIG", zap.Error(err))
		}

		if err := protojson.Unmarshal([]byte(env), pc); err != nil {
			logger.Fatal("error loading config from RIG_PROXY_CONFIG", zap.Error(err))
		}
	} else {
		logger.Warn("no RIG_PROXY_CONFIG env provided")
	}

	var publicKey interface{}
	var issuer string
	switch v := pc.GetJwtMethod().GetMethod().(type) {
	case nil:
	case *proto_proxy.JWTMethod_Certificate:
		p, _ := pem.Decode([]byte(v.Certificate))
		cert, err := x509.ParseCertificate(p.Bytes)
		if err != nil {
			logger.Fatal("error decoding certificate", zap.Error(err))
		}

		publicKey = cert.PublicKey
		issuer = cert.Issuer.CommonName
	case *proto_proxy.JWTMethod_Secret:
		publicKey = []byte(v.Secret)
	}

	logger.Info("loaded RIG_PROXY_CONFIG env", zap.Any("service_config", pc))

	for _, e := range pc.GetInterfaces() {
		target := fmt.Sprint(pc.GetTargetHost(), ":", e.GetTargetPort())
		switch e.GetLayer() {
		case proto_proxy.Layer_LAYER_4:
			p := &tcpproxy.Proxy{}
			p.AddRoute(fmt.Sprint(":", e.GetSourcePort()), tcpproxy.To(target))
			if err := p.Start(); err != nil {
				logger.Fatal("err setting up tcp proxy", zap.Error(err))
			}

		case proto_proxy.Layer_LAYER_7:
			p, err := proxy.New(target, logger)
			if err != nil {
				logger.Fatal("err setting up proxy", zap.Error(err))
			}

			var h http.Handler = p

			pid, err := uuid.Parse(pc.GetProjectId())
			if err != nil {
				logger.Fatal("invalid project ID", zap.String("project_id", pc.GetProjectId()), zap.Error(err))
			}

			for _, m := range e.GetMiddlewares() {
				switch v := m.Kind.(type) {
				case *capsule.Middleware_Authentication:
					h = &authenticationMiddleware{
						a:         v.Authentication,
						projectID: pid,
						publicKey: publicKey,
						issuer:    issuer,
						next:      h,
						logger:    logger,
					}
				default:
					logger.Fatal("invalid middleware", zap.Any("kind", reflect.TypeOf(v)))
				}
			}

			s := &http.Server{
				Addr:    fmt.Sprint(":", e.GetSourcePort()),
				Handler: h2c.NewHandler(h, &http2.Server{}),
			}

			logger.Info("starting service router", zap.Uint32("source_port", e.GetSourcePort()), zap.String("target", target))
			go func() {
				logger.Fatal("error listening", zap.Error(s.ListenAndServe()))
			}()

		default:
			logger.Fatal("invalid network layer", zap.Stringer("layer", e.GetLayer()))
		}
	}

	for {
		time.Sleep(time.Second)
	}
}

type authenticationMiddleware struct {
	a         *capsule.Authentication
	projectID uuid.UUID
	publicKey interface{}
	issuer    string
	next      http.Handler
	logger    *zap.Logger
}

func (m *authenticationMiddleware) handlePrefix(prefix string, a *capsule.Auth, w http.ResponseWriter, r *http.Request) {
	m.logger.Debug("using path prefix", zap.String("prefix", prefix))
	rp := path.Clean(r.URL.Path)
	for _, h := range m.a.GetHttp() {
		pp := path.Join(prefix, h.GetPath())
		if rp == pp || (!h.GetExact() && strings.HasPrefix(rp, pp)) {
			a = h.GetAuth()
			break
		}
	}

	switch v := a.GetMethod().(type) {
	case *capsule.Auth_AllowAuthorized_:
		h, err := m.handleJWTAuth(r.Header)
		if err != nil {
			w.WriteHeader(errors.ToHTTP(err))
			w.Write([]byte(errors.MessageOf(err)))
			w.Write([]byte("\n"))
			return
		}
		r.Header = h
	case *capsule.Auth_AllowAny_:
		break
	default:
		m.logger.Warn("invalid auth method for path prefix", zap.Any("method", v), zap.String("prefix", prefix))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("invalid auth configuration"))
		w.Write([]byte("\n"))
		return
	}

	m.next.ServeHTTP(w, r)
}

func (m *authenticationMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if m.a.GetEnabled() {
		m.handlePrefix("/", m.a.GetDefault(), w, r)
	} else {
		m.next.ServeHTTP(w, r)
	}
}

func (m *authenticationMiddleware) handleJWTAuth(h http.Header) (http.Header, error) {
	ah := h.Get("Authorization")
	if !strings.HasPrefix(ah, "Bearer ") {
		m.logger.Debug("request is missing authorization bearer")
		return h, errors.UnauthenticatedErrorf("missing authorization bearer")
	}

	jwtToken := strings.TrimPrefix(ah, "Bearer ")

	c := &auth.RigClaims{}
	token, err := jwt.ParseWithClaims(
		jwtToken,
		c,
		func(token *jwt.Token) (interface{}, error) {
			return m.publicKey, nil
		},
	)
	if err != nil {
		return h, errors.UnauthenticatedErrorf("%v", err)
	}

	if !token.Valid {
		return h, errors.InvalidArgumentErrorf("invalid JWT token format")
	}

	if c.GetIssuer() != m.issuer {
		return h, errors.InvalidArgumentErrorf("invalid JWT issuer")
	}

	if c.GetProjectID() != m.projectID {
		m.logger.Info("invalid project ID", zap.Stringer("claims_project_id", c.GetProjectID()), zap.Stringer("service_project_id", m.projectID))
		return h, errors.UnauthenticatedErrorf("invalid JWT token")
	}

	h.Del("Authorization")
	h.Set("X-Rig-User-ID", c.GetSubject().String())
	return h, nil
}
