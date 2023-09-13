package capsule

import (
	"context"
	"os"
	"strconv"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/cmd_config"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
	"github.com/spf13/cobra"
)

func CapsuleCreate(ctx context.Context, cmd *cobra.Command, args []string, nc rig.Client, cfg *cmd_config.Config) error {
	var err error
	if name == "" {
		name, err = common.PromptInput("Capsule name:", common.ValidateSystemNameOpt)
		if err != nil {
			return err
		}
	}

	var init []*capsule.Change
	var image string
	var replicas int
	if interactive {
		if ok, err := common.PromptConfirm("Do you want to add an initial image?", true); err != nil {
			return err
		} else if ok {
			if image, err = common.PromptInput("Image:", common.ValidateImageOpt); err != nil {
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
				Args:                 []string{},
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
					cf := &capsule.ConfigFile{}

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
		var buildID string
		if res, err := nc.Capsule().CreateBuild(ctx, &connect.Request[capsule.CreateBuildRequest]{
			Msg: &capsule.CreateBuildRequest{
				CapsuleId: capsuleID.String(),
				Image:     image,
			},
		}); err != nil {
			return err
		} else {
			buildID = res.Msg.GetBuildId()
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
