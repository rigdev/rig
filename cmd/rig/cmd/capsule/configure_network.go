package capsule

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
	"gopkg.in/yaml.v3"
)

func CapsuleConfigureNetwork(ctx context.Context, cmd *cobra.Command, args []string, capsuleID CapsuleID, nc rig.Client) error {
	var err error
	if networkFile == "" {
		networkFile, err = common.PromptInput("Enter Network file path:", common.ValidateNonEmptyOpt)
		if err != nil {
			return err
		}
	}

	bs, err := os.ReadFile(networkFile)
	if err != nil {
		return errors.InvalidArgumentErrorf("errors reading network info: %v", err)
	}

	var raw interface{}
	if err := yaml.Unmarshal(bs, &raw); err != nil {
		log.Fatal(err)
	}

	if bs, err = json.Marshal(raw); err != nil {
		log.Fatal(err)
	}

	n := &capsule.Network{}
	if err := protojson.Unmarshal(bs, n); err != nil {
		log.Fatal(err)
	}

	cmd.Println(n.GetInterfaces()[0])

	if _, err := nc.Capsule().Deploy(ctx, &connect.Request[capsule.DeployRequest]{
		Msg: &capsule.DeployRequest{
			CapsuleId: capsuleID,
			Changes: []*capsule.Change{{
				Field: &capsule.Change_Network{
					Network: n,
				},
			}},
		},
	}); err != nil {
		return err
	}

	cmd.Println("Network configured successfully!")

	return nil
}
