package storage

import (
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-api/api/v1/storage"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
)

func (c Cmd) listProviders(cmd *cobra.Command, args []string) error {
	ctx := c.Ctx
	Pagination := &model.Pagination{
		Offset: uint32(offset),
		Limit:  uint32(limit),
	}

	resp, err := c.Rig.Storage().ListProviders(ctx, &connect.Request[storage.ListProvidersRequest]{
		Msg: &storage.ListProvidersRequest{
			Pagination: Pagination,
		},
	})
	if err != nil {
		return err
	}

	if outputJson {
		for _, u := range resp.Msg.GetProviders() {
			cmd.Println(common.ProtoToPrettyJson(u))
		}
		return nil
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{fmt.Sprintf("Providers (%d)", resp.Msg.GetTotal()), "Name", "ID", "Backend", "#Buckets"})
	for i, u := range resp.Msg.GetProviders() {
		typ, err := getProviderType(u.GetConfig())
		if err != nil {
			return err
		}
		t.AppendRow(table.Row{i + 1, u.GetName(), u.GetProviderId(), typ, len(u.GetBuckets())})
	}
	cmd.Println(t.Render())
	return nil
}

func getProviderType(p *storage.Config) (string, error) {
	switch p.GetConfig().(type) {
	case *storage.Config_S3:
		return "s3", nil
	case *storage.Config_Gcs:
		return "gcs", nil
	case *storage.Config_Minio:
		return "minio", nil
	default:
		return "", errors.InvalidArgumentErrorf("unknown provider type")
	}
}
