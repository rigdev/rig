package tunnel

import (
	"fmt"
	"net"
	"strconv"
	"sync"
	"sync/atomic"

	"connectrpc.com/connect"
	api_tunnel "github.com/rigdev/rig-go-api/api/v1/tunnel"
	platformv1 "github.com/rigdev/rig-go-api/platform/v1"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/tunnel"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Server struct {
	logger  *zap.Logger
	handler *tunnelHandler
}

func New(logger *zap.Logger) *Server {
	s := &Server{
		logger: logger,
		handler: &tunnelHandler{
			logger:         logger,
			hostInterfaces: map[uint32]*platformv1.ProxyInterface{},
		},
	}

	return s
}

type tunnelHandler struct {
	api_tunnel.UnimplementedServiceServer

	listeners      []net.Listener
	hostInterfaces map[uint32]*platformv1.ProxyInterface
	logger         *zap.Logger
}

type tunnelScope struct {
	h      *tunnelHandler
	stream api_tunnel.Service_TunnelServer

	tunnelID atomic.Uint64

	lock    sync.Mutex
	tunnels map[uint64]*tunnel.Buffer
}

func (t *tunnelScope) Write(tunnelID uint64, data []byte) error {
	return t.stream.Send(&api_tunnel.TunnelResponse{
		Message: &api_tunnel.TunnelMessage{
			Message: &api_tunnel.TunnelMessage_Data{
				Data: &api_tunnel.TunnelData{
					TunnelId: tunnelID,
					Data:     data,
				},
			},
		},
	})
}

func (t *tunnelScope) Close(tunnelID uint64, err error) {
	logger := t.h.logger.With(zap.Uint64("tunnel_id", tunnelID))
	if err != nil {
		logger.Warn("socket closed with error", zap.Stringer("code", errors.CodeOf(err)), zap.String("message", errors.MessageOf(err)))
	} else {
		logger.Info("socket closed")
	}

	_ = t.stream.Send(&api_tunnel.TunnelResponse{
		Message: &api_tunnel.TunnelMessage{
			Message: &api_tunnel.TunnelMessage_Close{
				Close: &api_tunnel.TunnelClose{
					TunnelId: tunnelID,
					Code:     uint32(errors.CodeOf(err)),
					Message:  errors.MessageOf(err),
				},
			},
		},
	})
}

func (t *tunnelScope) NewTunnelID(port uint32) (uint64, *tunnel.Buffer, error) {
	tunnelID := t.tunnelID.Add(2)

	buff := tunnel.NewBuffer()

	t.lock.Lock()
	t.tunnels[tunnelID] = buff
	t.lock.Unlock()

	err := t.stream.Send(&api_tunnel.TunnelResponse{
		Message: &api_tunnel.TunnelMessage{
			Message: &api_tunnel.TunnelMessage_NewTunnel{
				NewTunnel: &api_tunnel.TunnelInfo{
					TunnelId: tunnelID,
					Port:     port,
				},
			},
		},
	})
	if err != nil {
		t.lock.Lock()
		delete(t.tunnels, tunnelID)
		t.lock.Unlock()
		return 0, nil, err
	}

	return tunnelID, buff, nil
}

func (t *tunnelScope) Target(tunnelID uint64, port uint32) (tunnel.Target, error) {
	cfg, ok := t.h.hostInterfaces[port]
	if !ok {
		return tunnel.Target{}, errors.NotFoundErrorf("tunnel port '%d' not found", port)
	}

	t.h.logger.Info("new outgoing request", zap.Uint64("tunnel_id", tunnelID), zap.String("target", cfg.GetTarget()))

	return tunnel.Target{
		Host:    cfg.GetTarget(),
		Options: cfg.GetOptions(),
	}, nil
}

func (h *tunnelHandler) Tunnel(stream api_tunnel.Service_TunnelServer) error {
	t := &tunnelScope{
		h:       h,
		stream:  stream,
		tunnels: map[uint64]*tunnel.Buffer{},
	}
	t.tunnelID.Store(1)

	for _, listener := range h.listeners {
		listener := listener

		_, portStr, err := net.SplitHostPort(listener.Addr().String())
		if err != nil {
			return err
		}

		port, err := strconv.ParseUint(portStr, 10, 32)
		if err != nil {
			return err
		}

		go func() {
			for {
				conn, err := listener.Accept()
				if err != nil {
					h.logger.Warn("error accepting connection", zap.Error(err))
					return
				}

				if stream.Context().Err() != nil {
					conn.Close()
					return
				}

				h.logger.Info("new incoming request", zap.Stringer("local", conn.LocalAddr()), zap.Stringer("remote", conn.RemoteAddr()))

				if err := tunnel.HandleInbound(stream.Context(), t, conn, uint32(port)); err != nil {
					h.logger.Error("error initializing reverse tunnel", zap.Error(err))
					conn.Close()
				}
			}
		}()
	}

	for {
		req, err := stream.Recv()
		if err != nil {
			return err
		}

		switch v := req.GetMessage().GetMessage().(type) {
		case *api_tunnel.TunnelMessage_NewTunnel:
			buff := tunnel.NewBuffer()

			t.lock.Lock()
			t.tunnels[v.NewTunnel.GetTunnelId()] = buff
			t.lock.Unlock()

			h.logger.Info("new incoming tunnel", zap.Uint64("tunnel_id", v.NewTunnel.GetTunnelId()))

			go tunnel.HandleOutbound(stream.Context(), t, v.NewTunnel.GetTunnelId(), v.NewTunnel.GetPort(), buff)

		case *api_tunnel.TunnelMessage_Data:
			t.lock.Lock()
			buff, ok := t.tunnels[v.Data.GetTunnelId()]
			t.lock.Unlock()
			if !ok {
				continue
			}

			if err := buff.Put(stream.Context(), v.Data.GetData()); err != nil {
				t.Close(v.Data.GetTunnelId(), err)
			}

		case *api_tunnel.TunnelMessage_Close:
			tunnelID := v.Close.GetTunnelId()
			t.lock.Lock()
			buff, ok := t.tunnels[tunnelID]
			if ok {
				delete(t.tunnels, tunnelID)
			}
			t.lock.Unlock()

			if ok {
				buff.Close()
				logger := h.logger.With(zap.Uint64("tunnel_id", tunnelID))
				if v.Close.Code == 0 {
					logger.Info("tunnel closed",
						zap.Stringer("code", connect.Code(v.Close.GetCode())),
						zap.String("message", v.Close.GetMessage()),
					)
				} else {
					logger.Warn("tunnel closed",
						zap.Uint64("tunnel_id", tunnelID),
						zap.Stringer("code", connect.Code(v.Close.GetCode())),
						zap.String("message", v.Close.GetMessage()),
					)
				}
			}
		}
	}
}

func (s *Server) AddCapsuleInterface(cfg *platformv1.ProxyInterface) error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GetPort()))
	if err != nil {
		return err
	}

	s.logger.Info("listening on interface", zap.String("addr", l.Addr().String()))

	s.handler.listeners = append(s.handler.listeners, l)

	return nil
}

func (s *Server) AddHostInterface(cfg *platformv1.ProxyInterface) error {
	s.handler.hostInterfaces[cfg.GetPort()] = cfg
	return nil
}

func (s *Server) Serve(port uint32) error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	s.logger.Info("listening on interface", zap.String("addr", l.Addr().String()))

	gs := grpc.NewServer()
	api_tunnel.RegisterServiceServer(gs, s.handler)
	return gs.Serve(l)
}
