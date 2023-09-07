package storage

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/jedib0t/go-pretty/v6/list"
	"github.com/rigdev/rig-go-api/api/v1/storage"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
)

func StorageUnlinkBucket(ctx context.Context, cmd *cobra.Command, args []string, nc rig.Client) error {
	l := list.NewWriter()
	l.SetStyle(list.StyleConnectedRounded)
	var bucket string
	var err error
	if len(args) < 1 {
		bucket, err = common.PromptGetInput("Bucket name:", ValidateBucketNameOpt)
		if err != nil {
			return err
		}
	} else {
		bucket = args[0]
	}

	_, err = nc.Storage().UnlinkBucket(ctx, &connect.Request[storage.UnlinkBucketRequest]{
		Msg: &storage.UnlinkBucketRequest{
			Bucket: bucket,
		},
	})
	if err != nil {
		return err
	}

	cmd.Println("Bucket unlinked: ", bucket)
	return nil
}
