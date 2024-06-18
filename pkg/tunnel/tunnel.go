package tunnel

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"

	platformv1 "github.com/rigdev/rig-go-api/platform/v1"
)

type Tunnel interface {
	Write(tunnelID uint64, data []byte) error
	Close(tunnelID uint64, err error)
	Target(tunnelID uint64, port uint32) (Target, error)
	NewTunnelID(port uint32) (uint64, *Buffer, error)
}

type Target struct {
	Host    string
	Options *platformv1.InterfaceOptions
}

// HandleInbound maintains a remote connection using the Tunnel interface. The `port` is
// the port "identifier", used to identify the remote target.
func HandleInbound(ctx context.Context, t Tunnel, conn net.Conn, port uint32) error {
	tunnelID, buff, err := t.NewTunnelID(port)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for {
			bs, err := buff.Take(ctx)
			if err == io.EOF {
				if v, ok := conn.(*net.TCPConn); ok {
					_ = v.CloseWrite()
				} else {
					conn.Close()
				}
				return
			} else if err != nil {
				t.Close(tunnelID, err)
				conn.Close()
				return
			}

			if _, err := conn.Write(bs); err != nil {
				t.Close(tunnelID, err)
				conn.Close()
				return
			}
		}
	}()

	go func() {
		defer wg.Done()

		for {
			bs := make([]byte, 1024)
			n, err := conn.Read(bs)
			if err == io.EOF {
				t.Close(tunnelID, nil)
				return
			} else if err != nil {
				t.Close(tunnelID, err)
				return
			}

			if err := t.Write(tunnelID, bs[:n]); err != nil {
				t.Close(tunnelID, err)
				return
			}
		}
	}()

	go func() {
		wg.Wait()
		conn.Close()
		t.Close(tunnelID, nil)
	}()

	return nil
}

// HandleOutbound creates a new outbound connection (TCP or HTTP) based on the Target matching the port "identifier".
func HandleOutbound(ctx context.Context, t Tunnel, tunnelID uint64, port uint32, buff *Buffer) {
	t.Close(tunnelID, handleOutboundInner(ctx, t, tunnelID, port, buff))
}

func handleOutboundInner(ctx context.Context, t Tunnel, tunnelID uint64, port uint32, buff *Buffer) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	target, err := t.Target(tunnelID, port)
	if err != nil {
		return err
	}

	var conn net.Conn

	if target.Options.GetTcp() {
		var d net.Dialer
		tcpConn, err := d.DialContext(ctx, "tcp", target.Host)
		if err != nil {
			return err
		}

		conn = tcpConn
	} else {
		t := url.URL{
			Scheme: "http",
			Host:   target.Host,
		}
		rp := &httputil.ReverseProxy{
			Rewrite: func(r *httputil.ProxyRequest) {
				r.SetURL(&t)

				if !target.Options.GetChangeOrigin() {
					r.Out.Host = r.In.Host
				}

				for key, value := range target.Options.GetHeaders() {
					r.Out.Header.Set(key, value)
				}
			},
			ModifyResponse: func(r *http.Response) error {
				if target.Options.GetAllowOrigin() != "" {
					r.Header.Set("Access-Control-Allow-Origin", target.Options.GetAllowOrigin())
				}

				return nil
			},
		}

		c1, c2 := net.Pipe()
		l := newSingleConnListener(c2)
		go func() {
			_ = http.Serve(l, rp)
		}()
		conn = c1
	}

	readDone := make(chan struct{})
	go func() {
		defer close(readDone)
		for {
			bs := make([]byte, 1024)
			n, err := conn.Read(bs)
			if err == io.EOF || err == io.ErrClosedPipe {
				return
			} else if err != nil {
				t.Close(tunnelID, err)
				cancel()
				return
			}

			if err := t.Write(tunnelID, bs[:n]); err != nil {
				t.Close(tunnelID, err)
				cancel()
				return
			}
		}
	}()

	defer conn.Close()

	for {
		bs, err := buff.Take(ctx)
		if err == io.EOF {
			if v, ok := conn.(*net.TCPConn); ok {
				_ = v.CloseWrite()
			} else {
				conn.Close()
			}
			// Wait for read done.
			select {
			case <-readDone:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		} else if err != nil {
			return err
		}

		if _, err := conn.Write(bs); err != nil {
			return err
		}
	}
}

type singleConnListener struct {
	ch chan net.Conn
}

func newSingleConnListener(conn net.Conn) *singleConnListener {
	l := &singleConnListener{
		ch: make(chan net.Conn, 1),
	}
	l.ch <- conn
	close(l.ch)
	return l
}

func (l *singleConnListener) Accept() (net.Conn, error) {
	conn, ok := <-l.ch
	if !ok {
		return nil, io.ErrClosedPipe
	}
	return conn, nil
}

func (l *singleConnListener) Close() error {
	return nil
}

func (l *singleConnListener) Addr() net.Addr {
	return noneAddr{}
}

type noneAddr struct{}

func (noneAddr) Network() string {
	return "none"
}

func (noneAddr) String() string {
	return "none"
}
