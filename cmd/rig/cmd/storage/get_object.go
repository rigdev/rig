package storage

import (
	"github.com/bufbuild/connect-go"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rigdev/rig-go-api/api/v1/storage"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
)

func (c Cmd) getObject(cmd *cobra.Command, args []string) error {
	ctx := c.Ctx
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
		res, err := c.Rig.Storage().GetObject(ctx, &connect.Request[storage.GetObjectRequest]{
			Msg: &storage.GetObjectRequest{
				Bucket: bucket,
				Path:   prefix,
			},
		})
		if err != nil {
			return err
		}

		if outputJson {
			cmd.Println(common.ProtoToPrettyJson(res.Msg.GetObject()))
			return nil
		}

		t := table.NewWriter()
		t.AppendHeader(table.Row{"Attribute", "Value"})
		t.AppendRows([]table.Row{
			{"Name", res.Msg.GetObject().GetPath()},
			{"Content type", res.Msg.GetObject().GetContentType()},
			{"Etag", res.Msg.GetObject().GetEtag()},
			{"Size", res.Msg.GetObject().GetSize()},
			{"Uploaded at", res.Msg.GetObject().GetLastModified().AsTime().Format("2006-01-02 15:04:05")},
		})
		cmd.Println(t.Render())

	} else {
		return errors.InvalidArgumentErrorf("invalid path: %s", path)
	}
	return nil
}
