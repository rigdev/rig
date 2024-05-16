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

func (c *Cmd) list(ctx context.Context, _ *cobra.Command, _ []string) error {
	resp, err := c.Rig.Cluster().List(ctx, connect.NewRequest(&cluster.ListRequest{}))
	if err != nil {
		return err
	}

	var clusters []string
	for _, c := range resp.Msg.GetClusters() {
		clusters = append(clusters, c.GetClusterId())
	}

	if flags.Flags.OutputType != common.OutputTypePretty {
		return common.FormatPrint(clusters, flags.Flags.OutputType)
	}

	// Yes, pretty-printing is also just yaml for this one
	bytes, err := yaml.Marshal(clusters)
	if err != nil {
		return err
	}

	fmt.Println(string(bytes))

	return nil
}
