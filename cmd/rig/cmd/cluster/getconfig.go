package cluster

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/cluster"
	"github.com/rigdev/rig-go-sdk"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func GetConfig(ctx context.Context, cmd *cobra.Command, client rig.Client) error {
	resp, err := client.Cluster().GetConfig(ctx, connect.NewRequest(&cluster.GetConfigRequest{}))
	if err != nil {
		return err
	}
	config := resp.Msg
	bytes, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	fmt.Println(string(bytes))

	return nil
}
