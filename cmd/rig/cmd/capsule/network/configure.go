package network

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"connectrpc.com/connect"
	"github.com/fatih/color"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/cmd/common"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
	"gopkg.in/yaml.v3"
)

func (c *Cmd) configure(ctx context.Context, cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		if err := c.configureInteractive(ctx, capsule_cmd.CapsuleID); err != nil {
			return err
		}
		return nil
	}

	bs, err := os.ReadFile(args[0])
	if err != nil {
		return errors.InvalidArgumentErrorf("errors reading network info: %v", err)
	}

	var raw interface{}
	if err := yaml.Unmarshal(bs, &raw); err != nil {
		return err
	}

	if bs, err = json.Marshal(raw); err != nil {
		return err
	}

	n := &capsule.Network{}
	if err := protojson.Unmarshal(bs, n); err != nil {
		return err
	}

	req := &connect.Request[capsule.DeployRequest]{
		Msg: &capsule.DeployRequest{
			CapsuleId: capsule_cmd.CapsuleID,
			Changes: []*capsule.Change{{
				Field: &capsule.Change_Network{
					Network: n,
				},
			}},
			ProjectId:     flags.GetProject(c.Cfg),
			EnvironmentId: flags.GetEnvironment(c.Cfg),
		},
	}

	if err := capsule_cmd.Deploy(ctx, c.Rig, req, forceDeploy); err != nil {
		return err
	}

	cmd.Println("Network configured successfully!")

	return nil
}

func (c *Cmd) configureInteractive(ctx context.Context, capsuleID string) error {
	resp, err := c.Rig.Capsule().ListRollouts(ctx, connect.NewRequest(&capsule.ListRolloutsRequest{
		CapsuleId: capsuleID,
		Pagination: &model.Pagination{
			Offset:     0,
			Limit:      1,
			Descending: true,
		},
		ProjectId:     flags.GetProject(c.Cfg),
		EnvironmentId: flags.GetEnvironment(c.Cfg),
	}))
	if err != nil {
		return err
	}
	rollouts := resp.Msg.GetRollouts()
	if len(rollouts) == 0 {
		return errors.New("capsule has no rollouts")
	}

	network := rollouts[0].GetConfig().GetNetwork()
	if network == nil {
		rollouts[0].Config.Network = &capsule.Network{}
		network = rollouts[0].Config.Network
	}

	for {
		idx, _, err := common.PromptSelect("Choose", []string{
			"Add new interface",
			"Delete interface",
			"See interface",
			"Apply and finish",
		}, common.SelectDontShowResultOpt)
		if err != nil {
			return err
		}

		isDone := false
		switch idx {
		case 0:
			if err := addInterface(network); err != nil {
				return err
			}
		case 1:
			if err := deleteInterface(network); err != nil {
				return err
			}
		case 2:
			if err := seeInterface(network); err != nil {
				return err
			}
		case 3:
			isDone = true
		}
		if isDone {
			break
		}
	}

	req := connect.NewRequest(&capsule.DeployRequest{
		CapsuleId: capsuleID,
		Changes: []*capsule.Change{{
			Field: &capsule.Change_Network{
				Network: network,
			},
		}},
		ProjectId:     flags.GetProject(c.Cfg),
		EnvironmentId: flags.GetEnvironment(c.Cfg),
	})
	if err := capsule_cmd.Deploy(ctx, c.Rig, req, forceDeploy); err != nil {
		return err
	}

	fmt.Println("Network configured successfully!")

	return nil
}

func addInterface(network *capsule.Network) error {
	printBanner("Adding Interface")
	defer printBannerEnd()

	var names []string
	var ports []string
	for _, i := range network.GetInterfaces() {
		names = append(names, i.GetName())
		ports = append(ports, strconv.Itoa(int(i.GetPort())))
	}

	name, err := common.PromptInput("Name of interface:", common.ValidateAndOpt(
		common.ValidateSystemName,
		common.ValidateUnique(names),
	))
	if err != nil {
		return err
	}

	portStr, err := common.PromptInput("Port:", common.ValidateAndOpt(
		common.ValidatePort,
		common.ValidateUnique(ports),
	))
	if err != nil {
		return err
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return err
	}

	capsuleInterface := &capsule.Interface{
		Port: uint32(port),
		Name: name,
	}

	public, err := common.PromptConfirm("Make public:", false)
	if err != nil {
		return err
	}

	if !public {
		network.Interfaces = append(network.Interfaces, capsuleInterface)
		return nil
	}

	capsuleInterface.Public = &capsule.PublicInterface{
		Enabled: true,
		Method:  &capsule.RoutingMethod{},
	}

	idx, _, err := common.PromptSelect("Routing method type:", []string{"Ingress", "Loadbalancer"})
	if err != nil {
		return err
	}
	switch idx {
	case 0:
		host, err := common.PromptInput("Ingress host:")
		if err != nil {
			return err
		}
		capsuleInterface.Public.Method.Kind = &capsule.RoutingMethod_Ingress_{
			Ingress: &capsule.RoutingMethod_Ingress{
				Host: host,
			},
		}
	case 1:
		portStr, err := common.PromptInput("Loadbalancer port:", common.ValidatePortOpt)
		if err != nil {
			return err
		}
		loadBalancerPort, err := strconv.Atoi(portStr)
		if err != nil {
			return err
		}
		capsuleInterface.Public.Method.Kind = &capsule.RoutingMethod_LoadBalancer_{
			LoadBalancer: &capsule.RoutingMethod_LoadBalancer{
				Port: uint32(loadBalancerPort),
			},
		}
	}
	network.Interfaces = append(network.Interfaces, capsuleInterface)

	return nil
}

func seeInterface(network *capsule.Network) error {
	printBanner("See interface")
	defer printBannerEnd()

	if len(network.GetInterfaces()) == 0 {
		fmt.Println("Capsule has no interfaces")
		return nil
	}

	var names []string
	for _, i := range network.Interfaces {
		names = append(names, i.GetName())
	}
	idx, _, err := common.PromptSelect("Choose", names, common.SelectDontShowResultOpt)
	if err != nil {
		return err
	}

	capsuleInterface := network.GetInterfaces()[idx]
	bytes, err := yaml.Marshal(capsuleInterface)
	if err != nil {
		return err
	}
	fmt.Println(string(bytes))

	return nil
}

func deleteInterface(network *capsule.Network) error {
	printBanner("Delete interface")
	defer printBannerEnd()

	if len(network.GetInterfaces()) == 0 {
		fmt.Println("Capsule has no interfaces")
		return nil
	}

	names := []string{"Go back"}
	for _, i := range network.Interfaces {
		names = append(names, i.GetName())
	}
	idx, _, err := common.PromptSelect("Choose", names)
	if err != nil {
		return err
	}

	if idx == 0 {
		return nil
	}
	idx--

	var newInterfaces []*capsule.Interface
	for i, capsuleInterface := range network.GetInterfaces() {
		if i != idx {
			newInterfaces = append(newInterfaces, capsuleInterface)
		}
	}
	network.Interfaces = newInterfaces

	return nil
}

var bannerLength = 30

func printBanner(s string) {
	color.Cyan(makeBanner(s))
}

func printBannerEnd() {
	color.Cyan(strings.Repeat("-", bannerLength+2))
}

func makeBanner(s string) string {
	sideLength := (bannerLength - len(s)) / 2
	side1 := strings.Repeat("-", sideLength)
	side2 := strings.Repeat("-", bannerLength-len(s)-len(side1))
	return side1 + " " + s + " " + side2
}
