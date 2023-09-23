package storage

import (
	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/storage"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
)

func (c Cmd) deleteProvider(cmd *cobra.Command, args []string) error {
	ctx := c.Ctx
	identifier := ""
	if len(args) > 0 {
		identifier = args[0]
	}
	g, uid, err := common.GetStorageProvider(ctx, identifier, c.Rig)
	if err != nil {
		return err
	}

	_, err = c.Rig.Storage().DeleteProvider(ctx, &connect.Request[storage.DeleteProviderRequest]{
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
