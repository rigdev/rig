package network

import (
	"encoding/json"
	"os"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/cmd/common"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
	"gopkg.in/yaml.v3"
)

func (c Cmd) configure(cmd *cobra.Command, args []string) error {
	ctx := c.Ctx
	var err error
	networkFile := ""
	if len(args) == 0 {
		networkFile, err = common.PromptInput("Enter Network file path:", common.ValidateNonEmptyOpt)
		if err != nil {
			return err
		}
	} else {
		networkFile = args[0]
	}

	bs, err := os.ReadFile(networkFile)
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
		},
	}

	_, err = c.Rig.Capsule().Deploy(ctx, req)
	if errors.IsFailedPrecondition(err) && errors.MessageOf(err) == "rollout already in progress" {
		_, err = capsule_cmd.AbortAndDeploy(ctx, capsule_cmd.CapsuleID, c.Rig, req)
	}
	if err != nil {
		return err
	}

	cmd.Println("Network configured successfully!")

	return nil
}
