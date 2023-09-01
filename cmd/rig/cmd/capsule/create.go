package capsule

import (
	"context"
	"strconv"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/rigdev/rig/cmd/rig/cmd/utils"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
	"github.com/spf13/cobra"
)

func CapsuleCreate(ctx context.Context, cmd *cobra.Command, args []string, nc rig.Client, cfg *base.Config) error {
	var err error
	if name == "" {
		name, err = utils.PromptGetInput("Capsule name: ", utils.ValidateSystemName)
		if err != nil {
			return err
		}
	}

	var init []*capsule.Change
	var image string
	var replicas int
	if interactive {
		if ok, err := utils.PromptConfirm("Do you want to add an initial image", true); err != nil {
			return err
		} else if ok {
			if image, err = utils.PromptGetInput("Image: ", utils.ValidateImage); err != nil {
				return err
			}

			if ok, err := utils.PromptConfirm("Does the image listen to a port", true); err != nil {
				return err
			} else if ok {
				ifc := &capsule.Interface{
					Name: "default",
				}
				portStr, err := utils.PromptGetInput("Which port: ", utils.ValidateInt)
				if err != nil {
					return err
				}

				port, err := strconv.Atoi(portStr)
				if err != nil {
					return err
				}

				ifc.Port = uint32(port)

				if ok, err := utils.PromptConfirm("Do you want to make the port public available", false); err != nil {
					return err
				} else if ok {
					ifc.Public = &capsule.PublicInterface{
						Enabled: true,
						Method:  &capsule.RoutingMethod{},
					}
					i, _, err := utils.PromptSelect("Which method?", []string{"Load balancer (raw traffic routing)", "Ingress (HTTP/HTTPS routing)"}, false)
					if err != nil {
						return err
					}

					switch i {
					case 0:
						portStr, err := utils.PromptGetInput("What public port to use: ", utils.ValidateInt)
						if err != nil {
							return err
						}

						port, err := strconv.Atoi(portStr)
						if err != nil {
							return err
						}

						ifc.Public.Method.Kind = &capsule.RoutingMethod_LoadBalancer_{
							LoadBalancer: &capsule.RoutingMethod_LoadBalancer{
								Port: uint32(port),
							},
						}
					default:
						return errors.InvalidArgumentErrorf("invalid public routing method")
					}
				}

				init = append(init, &capsule.Change{
					Field: &capsule.Change_Network{
						Network: &capsule.Network{
							Interfaces: []*capsule.Interface{ifc},
						},
					},
				})
			}
		}
		replicasStr, err := utils.PromptGetInputWithDefault("Replicas: ", utils.ValidateInt, "1")
		if err != nil {
			return err
		}

		replicas, err = strconv.Atoi(replicasStr)
		if err != nil {
			return err
		}
	}

	res, err := nc.Capsule().Create(ctx, &connect.Request[capsule.CreateRequest]{
		Msg: &capsule.CreateRequest{
			Name: name,
		},
	})
	if err != nil {
		return err
	}

	capsuleID, err := uuid.Parse(res.Msg.GetCapsuleId())
	if err != nil {
		return err
	}

	if image != "" {
		buildID := uuid.New().String()
		if _, err := nc.Capsule().CreateBuild(ctx, &connect.Request[capsule.CreateBuildRequest]{
			Msg: &capsule.CreateBuildRequest{
				CapsuleId: capsuleID.String(),
				BuildId:   buildID,
				Image:     image,
			},
		}); err != nil {
			return err
		}

		init = append(init, &capsule.Change{
			Field: &capsule.Change_BuildId{
				BuildId: buildID,
			},
		})
	}

	if replicas > 0 {
		init = append(init, &capsule.Change{
			Field: &capsule.Change_Replicas{
				Replicas: uint32(replicas),
			},
		})
	}

	if len(init) > 0 {
		if _, err := nc.Capsule().Deploy(ctx, &connect.Request[capsule.DeployRequest]{
			Msg: &capsule.DeployRequest{
				CapsuleId: capsuleID.String(),
				Changes:   init,
			},
		}); err != nil {
			return err
		}
	}

	cmd.Printf("Created new capsule '%v'\n", name)
	return nil
}
