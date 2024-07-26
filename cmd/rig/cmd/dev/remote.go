package dev

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/model"
	platformv1 "github.com/rigdev/rig-go-api/platform/v1"
	"github.com/rigdev/rig-go-api/v1alpha2"
	"github.com/rigdev/rig-go-sdk"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/pkg/cli"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
)

func (c *Cmd) remote(ctx context.Context, _ *cobra.Command, _ []string) error {
	remoteCtx := c.Scope.GetCfg().GetContext(remoteContext)
	if remoteCtx == nil {
		return errors.NotFoundErrorf("remote context `%s` not found", remoteContext)
	}

	remoteClientOpts, err := cli.GetClientOptions(c.Scope.GetCfg(), remoteCtx)
	if err != nil {
		return err
	}

	remoteClient := rig.NewClient(remoteClientOpts...)

	res, err := remoteClient.Capsule().ListRollouts(ctx, &connect.Request[capsule.ListRolloutsRequest]{
		Msg: &capsule.ListRolloutsRequest{
			CapsuleId:     capsuleName,
			ProjectId:     remoteCtx.GetProject(),
			EnvironmentId: remoteCtx.GetEnvironment(),
			Pagination: &model.Pagination{
				Limit:      1,
				Descending: true,
			},
		},
	})
	if err != nil {
		return err
	}

	if len(res.Msg.GetRollouts()) == 0 {
		return fmt.Errorf("no active rollouts for capsule `%s` in environment `%s`", capsuleName, remoteCtx.GetEnvironment())
	}

	instanceID, err := capsule_cmd.GetCapsuleInstance(ctx, remoteClient, remoteCtx, capsuleName)
	if err != nil {
		return err
	}

	remoteSpec := res.Msg.GetRollouts()[0].GetSpec()

	if len(interfaces) == 0 {
		for _, netIf := range remoteSpec.GetInterfaces() {
			interfaces = append(interfaces, netIf.GetName())
		}
	}

	var capInterfaces []*v1alpha2.CapsuleInterface

outer:
	for _, ifName := range interfaces {
		for _, netIf := range remoteSpec.GetInterfaces() {
			if ifName == netIf.GetName() {
				capInterfaces = append(capInterfaces, netIf)
				continue outer
			}

			asPort, _ := strconv.ParseUint(ifName, 10, 32)
			if asPort > 0 && asPort == uint64(netIf.GetPort()) {
				capInterfaces = append(capInterfaces, netIf)
				continue outer
			}
		}

		return fmt.Errorf("unknown interface `%s`", ifName)
	}

	errChan := make(chan error, 3)

	hostCfg := &platformv1.HostCapsule{
		Name:        capsuleName,
		Project:     c.Scope.GetCurrentContext().GetProject(),
		Environment: c.Scope.GetCurrentContext().GetEnvironment(),
		Network:     &platformv1.HostNetwork{},
	}
	for _, capIf := range capInterfaces {
		l, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			return err
		}

		port := uint32(capIf.GetPort())
		hostCfg.Network.CapsuleInterfaces = append(hostCfg.Network.CapsuleInterfaces, &platformv1.ProxyInterface{
			Port:   port,
			Target: l.Addr().String(),
		})

		go func() {
			errChan <- capsule_cmd.PortForwardOnListener(
				ctx, remoteClient, remoteCtx, capsuleName, instanceID, l, port, true)
		}()
	}

	go func() {
		errChan <- c.createHostTunnel(ctx, hostCfg)
	}()

	return <-errChan
}
