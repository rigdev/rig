package cluster

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/cluster"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func (c *Cmd) get(ctx context.Context, _ *cobra.Command, _ []string) error {
	resp, err := c.Rig.Cluster().GetConfig(ctx, connect.NewRequest(&cluster.GetConfigRequest{
		EnvironmentId: c.Scope.GetCurrentContext().GetEnvironment(),
	}))
	if err != nil {
		return err
	}
	config := resp.Msg

	if flags.Flags.OutputType != common.OutputTypePretty {
		return common.FormatPrint(config, flags.Flags.OutputType)
	}

	// Yes, pretty-printing is also just yaml for this one
	bytes, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	fmt.Println(string(bytes))

	return nil
}
