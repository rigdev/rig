package proxy

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/mwitkow/grpc-proxy/proxy"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Resolver interface {
	Resolve(string) (string, error)
}

type Proxy struct {
	target string
	logger *zap.Logger
	gp     *grpc.Server
	h      http.Handler
}

func New(target string, logger *zap.Logger) (*Proxy, error) {
	gc, err := grpc.DialContext(
		context.Background(),
		fmt.Sprint("dns:///", target),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(fmt.Sprintf("http://%s", target))
	if err != nil {
		return nil, err
	}

	return &Proxy{
		target: target,
		logger: logger,
		gp:     proxy.NewProxy(gc),
		h:      httputil.NewSingleHostReverseProxy(u),
	}, nil
}

func (p *Proxy) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if req.ProtoMajor == 2 && strings.HasPrefix(req.Header.Get("Content-Type"), "application/grpc") {
		// It's a gRPC request, use the gRPC proxy.
		p.logger.Info("proxying request", zap.String("host", req.Host), zap.Stringer("from", req.URL))
		p.gp.ServeHTTP(res, req)
		return
	}

	p.h.ServeHTTP(res, req)
}
