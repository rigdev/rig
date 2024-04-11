package image

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-api/api/v1/image"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/cmd/common"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/spf13/cobra"
)

func (c *Cmd) list(ctx context.Context, cmd *cobra.Command, _ []string) error {
	resp, err := c.Rig.Image().List(ctx, connect.NewRequest(
		&image.ListRequest{
			CapsuleId: capsule_cmd.CapsuleID,
			Pagination: &model.Pagination{
				Offset:     uint32(offset),
				Limit:      uint32(limit),
				Descending: true,
			},
			ProjectId: flags.GetProject(c.Scope),
		}),
	)
	if err != nil {
		return err
	}

	if flags.Flags.OutputType != common.OutputTypePretty {
		return common.FormatPrint(resp.Msg.GetImages(), flags.Flags.OutputType)
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{fmt.Sprintf("Images (%d)", resp.Msg.GetTotal()), "Digest", "Age", "Created By"})
	for _, b := range resp.Msg.GetImages() {
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
