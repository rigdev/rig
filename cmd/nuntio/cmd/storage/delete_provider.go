package storage

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/storage"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/rig/cmd/utils"
	"github.com/spf13/cobra"
)

func StorageDeleteProvider(ctx context.Context, cmd *cobra.Command, args []string, nc rig.Client) error {
	identifier := ""
	if len(args) > 0 {
		identifier = args[0]
	}
	g, uid, err := utils.GetStorageProvider(ctx, identifier, nc)
	if err != nil {
		return err
	}

	_, err = nc.Storage().DeleteProvider(ctx, &connect.Request[storage.DeleteProviderRequest]{
		Msg: &storage.DeleteProviderRequest{
			ProviderId: uid,
		},
	})
	if err != nil {
		return err
	}

	cmd.Println("Provider deleted: ", g.GetName())
	return nil
}
