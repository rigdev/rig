package storage

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/jedib0t/go-pretty/v6/list"
	"github.com/rigdev/rig-go-api/api/v1/storage"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
)

func (c Cmd) ls(cmd *cobra.Command, args []string) error {
	ctx := c.Ctx
	l := list.NewWriter()
	l.SetStyle(list.StyleConnectedRounded)

	if len(args) == 0 {
		// List buckets.
		token := ""
		for {
			res, err := c.Rig.Storage().ListBuckets(ctx, &connect.Request[storage.ListBucketsRequest]{
				Msg: &storage.ListBucketsRequest{
					Token: token,
				},
			})
			if err != nil {
				return err
			}

			if outputJson {
				for _, b := range res.Msg.GetBuckets() {
					cmd.Println(common.ProtoToPrettyJson(b))
				}
				return nil
			}

			for _, b := range res.Msg.GetBuckets() {
				l.AppendItem(fmt.Sprint("rig://", b.GetName(), "/ - (created_at=", b.GetCreatedAt().AsTime().Format(time.RFC3339), ")"))
				cmd.Println(l.Render())
				l.Reset()
				l.SetStyle(list.StyleConnectedRounded)
			}

			token = res.Msg.GetToken()
			if token == "" {
				break
			}
		}

		return nil
	}

	bucket, prefix, err := parseRigUri(args[0])
	if err != nil {
		return err
	}

	l.AppendItem(fmt.Sprint("rig://", path.Join(bucket, prefix)))
	l.Indent()

	// List files.
	token := ""
	for {
		res, err := c.Rig.Storage().ListObjects(ctx, &connect.Request[storage.ListObjectsRequest]{
			Msg: &storage.ListObjectsRequest{
				Token:     token,
				Bucket:    bucket,
				Prefix:    prefix,
				Recursive: storageRecursive,
			},
		})

		if outputJson {
			for _, o := range res.Msg.GetResults() {
				cmd.Println(common.ProtoToPrettyJson(o))
			}
			token = res.Msg.GetToken()
			if token == "" {
				return nil
			}
			continue
		}

		if err != nil {
			return err
		}

		for _, r := range res.Msg.GetResults() {
			switch v := r.GetResult().(type) {
			case *storage.ListObjectsResponse_Result_Folder:
				path := path.Clean(path.Join("/", v.Folder))
				path = strings.TrimPrefix(path, prefix)
				l.AppendItem(path)
			case *storage.ListObjectsResponse_Result_Object:
				prefix = listItem(l, prefix, v.Object)
			}
		}

		token = res.Msg.GetToken()
		if token == "" {
			break
		}
	}

	cmd.Println(l.Render())
	return nil
}

func listItem(l list.Writer, prefix string, item *storage.Object) string {
	fullPath := path.Clean(path.Join("/", item.GetPath()))
	// First find longest common prefix.
	for !strings.HasPrefix(fullPath, prefix) {
		prefix = path.Dir(prefix)
		l.UnIndent()
	}

	uniqueSuffix := fullPath
	uniqueSuffix = strings.TrimPrefix(uniqueSuffix, prefix)

	d, b := path.Split(uniqueSuffix)

	for _, s := range strings.Split(d, "/") {
		if s == "" {
			continue
		}
		l.AppendItem(fmt.Sprint(s, "/"))
		l.Indent()
	}

	if item.ContentType != "" {
		l.AppendItem(fmt.Sprint(b, " - ", item.GetContentType()))
	} else {
		l.AppendItem(b)
	}

	return path.Dir(fullPath)
}
