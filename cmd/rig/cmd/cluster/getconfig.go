package cluster

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/cluster"
	"github.com/rigdev/rig/cmd/rig/cmd/base"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func (c *Cmd) get(ctx context.Context, cmd *cobra.Command, args []string) error {
	resp, err := c.Rig.Cluster().GetConfig(ctx, connect.NewRequest(&cluster.GetConfigRequest{}))
	if err != nil {
		return err
	}
	config := resp.Msg

	if base.Flags.OutputType != base.OutputTypePretty {
		return base.FormatPrint(config)
	}

	// Yes, pretty-printing is also just yaml for this one
	bytes, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	fmt.Println(string(bytes))

	return nil
}
