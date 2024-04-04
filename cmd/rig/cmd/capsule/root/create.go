package root

import (
	"context"
	"os"
	"strconv"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/api/v1/image"
	"github.com/rigdev/rig/cmd/common"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
)

func (c *Cmd) create(ctx context.Context, cmd *cobra.Command, _ []string) error {
	var err error
	if capsule_cmd.CapsuleID == "" {
		capsule_cmd.CapsuleID, err = common.PromptInput("Capsule name:", common.ValidateSystemNameOpt)
		if err != nil {
			return err
		}
	}

	var init []*capsule.Change
	var imageID string
	var replicas int
	if interactive {
		if ok, err := common.PromptConfirm("Do you want to add an initial image?", true); err != nil {
			return err
		} else if ok {
			if imageID, err = common.PromptInput("Image:", common.ValidateImageOpt); err != nil {
				return err
			}

			if ok, err := common.PromptConfirm("Does the image listen to a port?", true); err != nil {
				return err
			} else if ok {
				ifc := &capsule.Interface{
					Name: "default",
				}
				portStr, err := common.PromptInput("Which port:", common.ValidateIntOpt)
				if err != nil {
					return err
				}

				port, err := strconv.Atoi(portStr)
				if err != nil {
					return err
				}

				ifc.Port = uint32(port)

				if ok, err := common.PromptConfirm("Do you want to make the port public available?", false); err != nil {
					return err
				} else if ok {
					ifc.Public = &capsule.PublicInterface{
						Enabled: true,
						Method:  &capsule.RoutingMethod{},
					}
					options := []string{"Load balancer (raw traffic routing)", "Ingress (HTTP/HTTPS routing)"}
					i, _, err := common.PromptSelect("Which method?", options)
					if err != nil {
						return err
					}

					switch i {
					case 0:
						portStr, err := common.PromptInput("What public port to use:", common.ValidateIntOpt)
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
			cSettings := &capsule.ContainerSettings{
				EnvironmentVariables: map[string]string{},
			}

			if ok, err := common.PromptConfirm("Do you want to add a command", false); err != nil {
				return err
			} else if ok {

				cmdStr, err := common.PromptInput("Command:", common.ValidateNonEmptyOpt)
				if err != nil {
					return err
				}

				cSettings.Command = cmdStr

				for {
					ok, err := common.PromptConfirm("Do you want to add an argument", false)
					if err != nil {
						return err
					}

					if !ok {
						break
					}

					argStr, err := common.PromptInput("Argument:", common.ValidateNonEmptyOpt)
					if err != nil {
						return err
					}

					cSettings.Args = append(cSettings.Args, argStr)
				}
			}

			for {
				ok, err := common.PromptConfirm("Do you want to add an environment variable", false)
				if err != nil {
					return err
				}

				if !ok {
					break
				}

				keyStr, err := common.PromptInput("Key:", common.ValidateNonEmptyOpt)
				if err != nil {
					return err
				}

				valueStr, err := common.PromptInput("Value:", common.ValidateNonEmptyOpt)
				if err != nil {
					return err
				}

				cSettings.EnvironmentVariables[keyStr] = valueStr
			}

			init = append(init, &capsule.Change{
				Field: &capsule.Change_ContainerSettings{
					ContainerSettings: cSettings,
				},
			})

			if ok, err := common.PromptConfirm("Do you want add config files", false); err != nil {
				return err
			} else if ok {
				for {
					cf := &capsule.Change_ConfigFile{}

					mountPath, err := common.PromptInput("Mount path: ", common.ValidateAbsPathOpt)
					if err != nil {
						return err
					}

					cf.Path = mountPath

					filepath, err := common.PromptInput("File path: ", common.ValidateNonEmptyOpt)
					if err != nil {
						return err
					}

					// Open file and parse the content into the file struct
					content, err := os.ReadFile(filepath)
					if err != nil {
						cmd.Println("Error opening file: ", err)
						continue
					}

					// if content size i greater than 1mb retry
					if len(content) > 1024*1024 {
						cmd.Println("File size is too big, max 1mb")
						continue
					}

					cf.Content = content

					init = append(init, &capsule.Change{
						Field: &capsule.Change_SetConfigFile{
							SetConfigFile: cf,
						},
					})

					if ok, err := common.PromptConfirm("Do you want to add another file", false); err != nil {
						return err
					} else if !ok {
						break
					}
				}
			}
		}
		replicasStr, err := common.PromptInput("Replicas:", common.ValidateIntOpt, common.InputDefaultOpt("1"))
		if err != nil {
			return err
		}

		replicas, err = strconv.Atoi(replicasStr)
		if err != nil {
			return err
		}
	}

	res, err := c.Rig.Capsule().Create(ctx, &connect.Request[capsule.CreateRequest]{
		Msg: &capsule.CreateRequest{
			Name:      capsule_cmd.CapsuleID,
			ProjectId: flags.GetProject(c.Scope),
		},
	})
	if err != nil {
		return err
	}

	capsuleID := res.Msg.GetCapsuleId()

	if imageID != "" {
		res, err := c.Rig.Image().Add(ctx, connect.NewRequest(&image.AddRequest{
			CapsuleId: capsuleID,
			Image:     imageID,
			ProjectId: flags.GetProject(c.Scope),
		}))
		if err != nil {
			return err
		}

		init = append(init, &capsule.Change{
			Field: &capsule.Change_ImageId{
				ImageId: res.Msg.GetImageId(),
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

	req := &connect.Request[capsule.DeployRequest]{
		Msg: &capsule.DeployRequest{
			CapsuleId:     capsuleID,
			Changes:       init,
			ProjectId:     flags.GetProject(c.Scope),
			EnvironmentId: flags.GetEnvironment(c.Scope),
		},
	}

	if len(init) > 0 {
		_, err = c.Rig.Capsule().Deploy(ctx, req)
		if errors.IsFailedPrecondition(err) && errors.MessageOf(err) == "rollout already in progress" {
			if forceDeploy {
				_, err = capsule_cmd.AbortAndDeploy(ctx, c.Rig, req)
			} else {
				_, err = capsule_cmd.PromptAbortAndDeploy(ctx, c.Rig, req)
			}
		}
		if err != nil {
			return err
		}

	}

	cmd.Printf("Created new capsule '%v'\n", capsule_cmd.CapsuleID)
	return nil
}
