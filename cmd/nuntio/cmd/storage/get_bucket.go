package storage

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-api/api/v1/storage"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/rig/cmd/utils"
	"github.com/spf13/cobra"
)

func StorageGetBucket(ctx context.Context, cmd *cobra.Command, args []string, nc rig.Client) error {
	var bucket string
	var err error
	if len(args) < 1 {
		bucket, err = utils.PromptGetInput("Bucket name:", ValidateBucketName)
		if err != nil {
			return err
		}
	} else {
		bucket = args[0]
	}

	res, err := nc.Storage().GetBucket(ctx, &connect.Request[storage.GetBucketRequest]{
		Msg: &storage.GetBucketRequest{
			Bucket: bucket,
		},
	})
	if err != nil {
		return err
	}

	if outputJson {
		cmd.Println(utils.ProtoToPrettyJson(res.Msg.GetBucket()))
		return nil
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{"Attribute", "Value"})
	t.AppendRows([]table.Row{
		{"Name", res.Msg.GetBucket().GetName()},
		{"Provider name", res.Msg.GetBucket().GetProviderBucket()},
		{"Region", res.Msg.GetBucket().GetRegion()},
		{"Created", res.Msg.GetBucket().GetCreatedAt().AsTime().Format("2006-01-02 15:04:05")},
	})

	cmd.Println(t.Render())
	return nil
}
