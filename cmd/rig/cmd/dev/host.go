package dev

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	platformv1 "github.com/rigdev/rig-go-api/platform/v1"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/obj"
	"github.com/spf13/cobra"
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

	return c.createHostTunnel(ctx, cfg)
}
