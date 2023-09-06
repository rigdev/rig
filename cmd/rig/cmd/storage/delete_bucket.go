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

func StorageDeleteBucket(ctx context.Context, cmd *cobra.Command, args []string, nc rig.Client) error {
	l := list.NewWriter()
	l.SetStyle(list.StyleConnectedRounded)
	var bucket string
	var err error
	if len(args) < 1 {
		bucket, err = common.PromptGetInput("Bucket name:", ValidateBucketName)
		if err != nil {
			return err
		}
	} else {
		bucket = args[0]
	}

	_, err = nc.Storage().DeleteBucket(ctx, &connect.Request[storage.DeleteBucketRequest]{
		Msg: &storage.DeleteBucketRequest{
			Bucket: bucket,
		},
	})
	if err != nil {
		return err
	}

	cmd.Println("Bucket deleted: ", bucket)
	return nil
}
