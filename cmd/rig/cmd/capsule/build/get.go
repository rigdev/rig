package build

import (
	"context"
	"fmt"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/spf13/cobra"
)

func get(ctx context.Context, cmd *cobra.Command, nc rig.Client) error {
	resp, err := nc.Capsule().ListBuilds(ctx, &connect.Request[capsule.ListBuildsRequest]{
		Msg: &capsule.ListBuildsRequest{
			CapsuleId: capsule_cmd.CapsuleID,
			Pagination: &model.Pagination{
				Offset:     uint32(offset),
				Limit:      uint32(limit),
				Descending: true,
			},
		},
	})
	if err != nil {
		return err
	}

	if outputJSON {
		for _, b := range resp.Msg.GetBuilds() {
			cmd.Println(common.ProtoToPrettyJson(b))
		}
		return nil
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{fmt.Sprintf("Builds (%d)", resp.Msg.GetTotal()), "Digest", "Age", "Created By"})
	for _, b := range resp.Msg.GetBuilds() {
		t.AppendRow(table.Row{
			fmt.Sprint(b.GetRepository(), ":", b.GetTag()),
			capsule_cmd.TruncatedFixed(b.GetDigest(), 19),
			common.FormatDuration(time.Since(b.GetCreatedAt().AsTime())),
			b.GetCreatedBy().GetPrintableName(),
		})
	}
	cmd.Println(t.Render())

	return nil
}
