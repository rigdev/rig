package root

import (
	"fmt"

	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/model"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func makeRollout() *capsule.RolloutConfig {
	return &capsule.RolloutConfig{
		BuildId: "image",
		Network: &capsule.Network{
			Interfaces: []*capsule.Interface{
				{
					Port: 8080,
					Name: "http",
					Public: &capsule.PublicInterface{
						Enabled: true,
						Method: &capsule.RoutingMethod{
							Kind: &capsule.RoutingMethod_LoadBalancer_{
								LoadBalancer: &capsule.RoutingMethod_LoadBalancer{
									Port: 5000,
								},
							},
						},
					},
				},
				{
					Port: 8000,
					Name: "http",
					Public: &capsule.PublicInterface{
						Enabled: true,
						Method: &capsule.RoutingMethod{
							Kind: &capsule.RoutingMethod_LoadBalancer_{
								LoadBalancer: &capsule.RoutingMethod_LoadBalancer{
									Port: 5000,
								},
							},
						},
					},
				},
			},
		},
		ContainerSettings: &capsule.ContainerSettings{
			EnvironmentVariables: map[string]string{
				"VAR": "value",
			},
			Command: "cmd",
			Args:    []string{"arg1", "arg2"},
		},
		AutoAddRigServiceAccounts: true,
		ConfigFiles: []*capsule.ConfigFile{
			{
				Path:    "/etc/path/config.yaml",
				Content: []byte("some yaml"),
				UpdatedBy: &model.Author{
					Identifier:    "matias@rig.dev",
					PrintableName: "Matias Frank Jensen",
					Account: &model.Author_UserId{
						UserId: "matias@rig.dev",
					},
				},
				UpdatedAt: timestamppb.Now(),
				IsSecret:  false,
			},
		},
		HorizontalScale: &capsule.HorizontalScale{
			MaxReplicas: 10,
			MinReplicas: 5,
			CpuTarget: &capsule.CPUTarget{
				AverageUtilizationPercentage: 80,
			},
		},
	}
}

func (c Cmd) rolloutTest(cmd *cobra.Command, args []string) error {
	c1 := makeRollout()
	c2 := makeRollout()
	c2.HorizontalScale.CpuTarget.AverageUtilizationPercentage = 60
	c2.Network.Interfaces[1].Port = 6969
	c2.Network.Interfaces = append(c2.Network.Interfaces, &capsule.Interface{
		Port: 1234,
		Name: "hej",
	})
	c2.Network.Interfaces = append(c2.Network.Interfaces, &capsule.Interface{
		Port: 4000,
		Name: "hej2",
	})
	c2.ConfigFiles = nil
	c2.ContainerSettings.Args = []string{"somearg", "allnewargs", "afa", "c", "d"}
	c2.ContainerSettings.EnvironmentVariables["newvar"] = "hej"

	fmt.Println("========================= PROTO ==========================")
	diff := rolloutDiff(c1, c2)
	printRolloutConfigDiff(diff.Root, diff.Nodes)

	fmt.Println()
	fmt.Println("========================= JSON ==========================")
	jdiff, err := jsonDiff(c1, c2)
	if err != nil {
		return err
	}
	jdiff.Print()

	return nil
}
