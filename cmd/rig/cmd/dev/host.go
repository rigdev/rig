package dev

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rigdev/rig-go-api/api/v1/capsule"
	api_tunnel "github.com/rigdev/rig-go-api/api/v1/tunnel"
	platformv1 "github.com/rigdev/rig-go-api/platform/v1"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/obj"
	"github.com/rigdev/rig/pkg/tunnel"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func parseInterface(arg string) (*platformv1.ProxyInterface, error) {
	parts := strings.Split(arg, ",")

	base := parts[0]
	baseParts := strings.Split(base, ":")
	if len(baseParts) < 3 {
		return nil, errors.InvalidArgumentErrorf(
			"wrong format of format rule, expected `local-port:target-capsule:target-port`")
	}

	port, err := strconv.ParseUint(baseParts[0], 10, 32)
	if err != nil {
		return nil, errors.InvalidArgumentErrorf("invalid port '%s': %v", baseParts[0], err)
	}

	allowOrigin := ""
	tcp := false
	for _, opt := range parts[1:] {
		optParts := strings.SplitN(opt, "=", 2)
		switch optParts[0] {
		case "allow-origin":
			allowOrigin = optParts[1]
		case "tcp":
			tcp = true
		default:
			return nil, errors.InvalidArgumentErrorf("invalid option '%s'", optParts[0])
		}
	}

	return &platformv1.ProxyInterface{
		Port:   uint32(port),
		Target: baseParts[1] + ":" + baseParts[2],
		Options: &platformv1.InterfaceOptions{
			Tcp:         tcp,
			AllowOrigin: allowOrigin,
		},
	}, nil
}

func (c *Cmd) host(ctx context.Context, cmd *cobra.Command, _ []string) error {
	cfg := &platformv1.HostCapsule{}

	if cmd.Flags().Changed("path") {
		bs, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		if err := obj.Decode(bs, cfg); err != nil {
			return err
		}
	}

	if cfg.GetName() != "" {
		capsuleName = cfg.GetName()
	}

	if cfg.GetEnvironment() != "" {
		flags.Flags.Environment = cfg.GetEnvironment()
	}

	if cfg.GetProject() != "" {
		flags.Flags.Project = cfg.GetProject()
	}

	if len(capsuleName) == 0 {
		if !c.Scope.IsInteractive() {
			return errors.InvalidArgumentErrorf("missing capsule name flag")
		}

		name, err := capsule_cmd.SelectCapsule(ctx, c.Rig, c.Prompter, c.Scope)
		if err != nil {
			return err
		}

		capsuleName = name
	}

	if cfg.GetNetwork() == nil {
		cfg.Network = &platformv1.HostNetwork{}
	}

	for _, arg := range capsuleInterface {
		proxyIf, err := parseInterface(arg)
		if err != nil {
			return err
		}

		cfg.Network.CapsuleInterfaces = append(cfg.Network.CapsuleInterfaces, proxyIf)
	}

	for _, arg := range hostInterface {
		proxyIf, err := parseInterface(arg)
		if err != nil {
			return err
		}

		cfg.Network.HostInterfaces = append(cfg.Network.HostInterfaces, proxyIf)
	}

	if printConfig {
		bs, err := obj.EncodeAny(cfg)
		if err != nil {
			return err
		}

		fmt.Println(string(bs))
		return nil
	}

	if cfg.GetNetwork().GetTunnelPort() == 0 {
		cfg.Network.TunnelPort = 10042
	}

	bs, err := obj.EncodeAny(cfg)
	if err != nil {
		return err
	}

	var changes []*capsule.Change
	changes = append(changes, &capsule.Change{
		Field: &capsule.Change_SetConfigFile{
			SetConfigFile: &capsule.Change_ConfigFile{
				Path:    "/capsule.yaml",
				Content: bs,
			},
		},
	})

	var interfaces []*capsule.Interface
	for _, capIf := range cfg.GetNetwork().GetCapsuleInterfaces() {
		interfaces = append(interfaces, &capsule.Interface{
			Port: capIf.Port,
			Name: fmt.Sprintf("forward-%d", capIf.Port),
		})
	}

	if cfg.GetNetwork().GetTunnelPort() != 0 {
		interfaces = append(interfaces, &capsule.Interface{
			Port: cfg.GetNetwork().GetTunnelPort(),
			Name: "host-tunnel",
		})
	}

	changes = append(changes, &capsule.Change{
		Field: &capsule.Change_Network{
			Network: &capsule.Network{
				Interfaces: interfaces,
			},
		},
	})

	changes = append(changes, &capsule.Change{Field: &capsule.Change_AddImage_{
		AddImage: &capsule.Change_AddImage{
			Image: "ghcr.io/rigdev/rig-proxy:dev",
		},
	}})

	revision, rolloutID, err := capsule_cmd.Deploy(ctx, c.Rig, c.Scope, capsuleName, changes, true, false, 0)
	if err != nil {
		return err
	}

	if err := capsule_cmd.WaitForRollout(ctx, c.Rig, c.Scope, capsuleName, revision, rolloutID); err != nil {
		return err
	}

	instanceID := ""

	if instanceID == "" {
		if instanceID, err = capsule_cmd.GetCapsuleInstance(ctx, c.Rig, c.Scope, capsuleName); err != nil {
			return err
		}
	}

	capInterfaces := map[uint32]*platformv1.ProxyInterface{}
	for _, capIf := range cfg.GetNetwork().GetCapsuleInterfaces() {
		capInterfaces[capIf.GetPort()] = capIf
	}

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}

	gc, err := grpc.NewClient(l.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}

	ls := map[uint32]hostListener{}
	for _, hostIf := range cfg.GetNetwork().GetHostInterfaces() {
		l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", hostIf.GetPort()))
		if err != nil {
			return err
		}

		_, portStr, err := net.SplitHostPort(l.Addr().String())
		if err != nil {
			return err
		}

		port, err := strconv.ParseUint(portStr, 10, 32)
		if err != nil {
			return err
		}

		ls[uint32(port)] = hostListener{
			cfg:      hostIf,
			listener: l,
		}
	}

	tunnelClient := api_tunnel.NewServiceClient(gc)
	go func() {
		for {

			tunnelStream, err := tunnelClient.Tunnel(ctx)
			if err != nil {
				fmt.Println("[rig] error establishing tunnel: ", err)
				time.Sleep(1 * time.Second)
				continue
			}

			rt := &clientTunnel{
				tunnelStream:   tunnelStream,
				capInterfaces:  capInterfaces,
				tunnels:        map[uint64]*tunnel.Buffer{},
				hostInterfaces: ls,
			}

			if err := rt.Run(ctx); err != nil {
				fmt.Println("[rig] err processing tunnel: ", err)
			}

			time.Sleep(1 * time.Second)
		}
	}()

	return capsule_cmd.PortForwardOnListener(
		ctx, c.Rig, c.Scope, capsuleName, instanceID, l, cfg.GetNetwork().GetTunnelPort(), true)
}

type hostListener struct {
	cfg      *platformv1.ProxyInterface
	listener net.Listener
}

type clientTunnel struct {
	tunnelStream   api_tunnel.Service_TunnelClient
	capInterfaces  map[uint32]*platformv1.ProxyInterface
	hostInterfaces map[uint32]hostListener

	tunnelID atomic.Uint64
	lock     sync.Mutex
	tunnels  map[uint64]*tunnel.Buffer
}

func (t *clientTunnel) Run(ctx context.Context) error {
	for port, listener := range t.hostInterfaces {
		port := port
		listener := listener
		go func() {
			for {
				conn, err := listener.listener.Accept()
				if err != nil {
					fmt.Println("[rig] error accepting connection:", err)
					return
				}

				fmt.Printf("[rig] new incoming request %s -> %s\n", conn.LocalAddr(), listener.cfg.GetTarget())

				if err := tunnel.HandleInbound(ctx, t, conn, port); err != nil {
					fmt.Println("[rig] error initializing reverse tunnel: ", err)
				}
			}
		}()
	}

	for {
		res, err := t.tunnelStream.Recv()
		if err != nil {
			return err
		}

		switch v := res.GetMessage().GetMessage().(type) {
		case *api_tunnel.TunnelMessage_NewTunnel:
			tunnelID := v.NewTunnel.GetTunnelId()
			buff := tunnel.NewBuffer()

			t.lock.Lock()
			t.tunnels[tunnelID] = buff
			t.lock.Unlock()

			go tunnel.HandleOutbound(ctx, t, tunnelID, v.NewTunnel.GetPort(), buff)

		case *api_tunnel.TunnelMessage_Data:
			t.lock.Lock()
			buff, ok := t.tunnels[v.Data.GetTunnelId()]
			t.lock.Unlock()
			if !ok {
				continue
			}

			if err := buff.Put(ctx, v.Data.GetData()); err != nil {
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
			}
		}
	}
}

func (t *clientTunnel) Write(tunnelID uint64, data []byte) error {
	return t.tunnelStream.Send(&api_tunnel.TunnelRequest{
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

func (t *clientTunnel) Close(tunnelID uint64, err error) {
	_ = t.tunnelStream.Send(&api_tunnel.TunnelRequest{
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

func (t *clientTunnel) Target(_ uint64, port uint32) (tunnel.Target, error) {
	cfg, ok := t.capInterfaces[port]
	if !ok {
		return tunnel.Target{}, errors.NotFoundErrorf("tunnel port '%d' not found", port)
	}

	return tunnel.Target{Host: cfg.GetTarget(), Options: cfg.GetOptions()}, nil
}

func (t *clientTunnel) NewTunnelID(port uint32) (uint64, *tunnel.Buffer, error) {
	tunnelID := t.tunnelID.Add(2)

	buff := tunnel.NewBuffer()

	t.lock.Lock()
	t.tunnels[tunnelID] = buff
	t.lock.Unlock()

	err := t.tunnelStream.Send(&api_tunnel.TunnelRequest{
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
