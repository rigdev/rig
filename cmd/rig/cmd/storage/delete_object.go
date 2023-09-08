package storage

import (
	"context"

	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/pkg/errors"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/storage"
	"github.com/spf13/cobra"
)

func StorageDeleteObject(ctx context.Context, cmd *cobra.Command, args []string, nc rig.Client) error {
	var path string
	var err error
	if len(args) < 1 {
		path, err = common.PromptInput("Object path:", common.ValidateNonEmptyOpt)
		if err != nil {
			return err
		}
	} else {
		path = args[0]
	}
	if isRigUri(path) {
		bucket, prefix, err := parseRigUri(path)
		if err != nil {
			return err
		}
		_, err = nc.Storage().DeleteObject(ctx, &connect.Request[storage.DeleteObjectRequest]{
			Msg: &storage.DeleteObjectRequest{
				Bucket: bucket,
				Path:   prefix,
			},
		})
		if err != nil {
			return err
		}
	} else {
		return errors.InvalidArgumentErrorf("invalid path: %s", path)
	}

	cmd.Println("Object deleted at: ", path)
	return nil
}
