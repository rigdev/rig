package root

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/model"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/durationpb"
)

func (c *Cmd) portForward(
	ctx context.Context,
	_ *cobra.Command,
	args []string,
) error {
	capsuleID := capsule_cmd.CapsuleID

	res, err := c.Rig.Capsule().ListRollouts(ctx, &connect.Request[capsule.ListRolloutsRequest]{
		Msg: &capsule.ListRolloutsRequest{
			ProjectId:     c.Scope.GetCurrentContext().GetProject(),
			EnvironmentId: c.Scope.GetCurrentContext().GetEnvironment(),
			CapsuleId:     capsuleID,
			Pagination: &model.Pagination{
				Descending: true,
				Limit:      1,
			},
		},
	})
	if err != nil {
		return err
	}

	if len(res.Msg.GetRollouts()) == 0 {
		return fmt.Errorf("capsule %s is not running", capsuleID)
	}

	spec := res.Msg.GetRollouts()[0].GetSpec()

	if len(spec.GetInterfaces()) == 0 {
		return fmt.Errorf("capsule has no network interfaces")
	}

	var localPort uint32
	var remotePort uint32
	if len(args) > 1 {
		parts := strings.SplitN(args[1], ":", 2)
		hasLocalPort := len(parts) == 2

		interfaceName := parts[0]
		if hasLocalPort {
			interfaceName = parts[1]
			port, err := strconv.ParseUint(parts[0], 10, 32)
			if err != nil {
				return err
			}

			localPort = uint32(port)
		}

		for _, i := range spec.GetInterfaces() {
			if i.GetName() == interfaceName || strconv.Itoa(int(i.GetPort())) == interfaceName {
				remotePort = uint32(i.GetPort())
				break
			}
		}

		if remotePort == 0 {
			return fmt.Errorf("no network interface matching '%s'", interfaceName)
		}

		if !hasLocalPort {
			localPort = remotePort
		}
	} else {
		if len(spec.GetInterfaces()) > 1 {
			var choices []string
			for _, i := range spec.GetInterfaces() {
				choices = append(choices, fmt.Sprintf("%s (port %d)", i.GetName(), i.GetPort()))
			}

			i, _, err := c.Prompter.Select("Select a network interface", choices)
			if err != nil {
				return err
			}

			remotePort = uint32(spec.GetInterfaces()[i].GetPort())
		} else {
			remotePort = uint32(spec.GetInterfaces()[0].GetPort())
		}

		localPort = remotePort
	}

	if instanceID == "" {
		instancesRes, err := c.Rig.Capsule().ListInstances(ctx, &connect.Request[capsule.ListInstancesRequest]{
			Msg: &capsule.ListInstancesRequest{
				ProjectId:     c.Scope.GetCurrentContext().GetProject(),
				EnvironmentId: c.Scope.GetCurrentContext().GetEnvironment(),
				CapsuleId:     capsuleID,
				Pagination: &model.Pagination{
					Limit: 1,
				},
			},
		})
		if err != nil {
			return err
		}

		if len(instancesRes.Msg.Instances) == 0 {
			return errors.NotFoundErrorf("no instances found for capsule")
		}

		instanceID = instancesRes.Msg.Instances[0].GetInstanceId()
	}

	l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", localPort))
	if err != nil {
		return err
	}

	fmt.Printf("[rig] connected to instance '%s', accepting traffic on %s\n", instanceID, l.Addr().String())

	if follow {
		go func() {
			for {
				time.Sleep(1 * time.Second)

				res, err := c.Rig.Capsule().Logs(ctx, &connect.Request[capsule.LogsRequest]{
					Msg: &capsule.LogsRequest{
						ProjectId:     c.Scope.GetCurrentContext().GetProject(),
						EnvironmentId: c.Scope.GetCurrentContext().GetEnvironment(),
						CapsuleId:     capsuleID,
						InstanceId:    instanceID,
						Follow:        true,
						Since:         durationpb.New(10 * time.Second),
					},
				})
				if err != nil {
					fmt.Printf("[rig] error tailing logs: %v\n", err)
					continue
				}

				for res.Receive() {
					switch v := res.Msg().GetLog().GetMessage().GetMessage().(type) {
					case *capsule.LogMessage_ContainerTermination_:
						fmt.Printf("[rig] instance restarted")
						res.Close()
					case *capsule.LogMessage_Stdout:
						os.Stdout.Write(v.Stdout)
					case *capsule.LogMessage_Stderr:
						os.Stdout.Write(v.Stderr)
					}
				}

				if err := res.Err(); err != nil {
					fmt.Printf("[rig] error tailing logs: %v\n", err)
				}

				res.Close()
			}
		}()
	}

	return capsule_cmd.PortForwardOnListener(
		ctx, c.Rig, c.Scope.GetCurrentContext(), capsuleID, instanceID, l, remotePort, verbose)
}
