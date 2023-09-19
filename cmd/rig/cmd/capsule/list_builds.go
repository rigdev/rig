package capsule

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
)

func CapsuleListBuilds(ctx context.Context, cmd *cobra.Command, capsuleID CapsuleID, nc rig.Client) error {
	resp, err := nc.Capsule().ListBuilds(ctx, &connect.Request[capsule.ListBuildsRequest]{
		Msg: &capsule.ListBuildsRequest{
			CapsuleId: capsuleID,
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
	t.AppendHeader(table.Row{fmt.Sprintf("Builds (%d)", resp.Msg.GetTotal()), "Tag", "Digest", "Age", "Created By"})
	for _, b := range resp.Msg.GetBuilds() {
		t.AppendRow(table.Row{
			fmt.Sprint(b.GetRepository(), ":", b.GetTag()),
			truncatedFixed(b.GetDigest(), 19),
			time.Since(b.GetCreatedAt().AsTime()).Truncate(time.Second),
			b.GetCreatedBy().GetPrintableName(),
		})
	}
	cmd.Println(t.Render())

	return nil
}

func truncated(str string, max int) string {
	if len(str) > max {
		return str[:strings.LastIndexAny(str[:max], " .,:;-")] + "..."
	}

	return str
}

func truncatedFixed(str string, max int) string {
	if len(str) > max {
		return str[:max] + "..."
	}

	return str
}
