package storage

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-api/api/v1/storage"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/rig/cmd/utils"
	storage_service "github.com/rigdev/rig/internal/service/storage"
	"github.com/spf13/cobra"
)

func StorageListProviders(ctx context.Context, cmd *cobra.Command, nc rig.Client) error {
	Pagination := &model.Pagination{
		Offset: uint32(offset),
		Limit:  uint32(limit),
	}

	resp, err := nc.Storage().ListProviders(ctx, &connect.Request[storage.ListProvidersRequest]{
		Msg: &storage.ListProvidersRequest{
			Pagination: Pagination,
		},
	})
	if err != nil {
		return err
	}

	if outputJson {
		for _, u := range resp.Msg.GetProviders() {
			cmd.Println(utils.ProtoToPrettyJson(u))
		}
		return nil
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{fmt.Sprintf("Providers (%d)", resp.Msg.GetTotal()), "Name", "ID", "Backend", "#Buckets"})
	for i, u := range resp.Msg.GetProviders() {
		typ, err := storage_service.GetProviderType(u.GetConfig())
		if err != nil {
			return err
		}
		t.AppendRow(table.Row{i + 1, u.GetName(), u.GetProviderId(), typ, len(u.GetBuckets())})
	}
	cmd.Println(t.Render())
	return nil
}
