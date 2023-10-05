package gcs

import (
	"context"
	"io"
	"strings"

	gStorage "cloud.google.com/go/storage"
	"github.com/rigdev/rig-go-api/api/v1/storage"
	"github.com/rigdev/rig/pkg/iterator"
	gIterator "google.golang.org/api/iterator"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Storage) ListObjects(ctx context.Context, bucketName, token, prefix,
	startpath, endpath string, recursive bool, limit uint32,
) (string, iterator.Iterator[*storage.ListObjectsResponse_Result], error) {
	prefix = strings.TrimPrefix(prefix, "/")

	if token != "" {
		startpath = token
	}

	q := &gStorage.Query{
		Prefix:      prefix,
		StartOffset: startpath,
		EndOffset:   endpath,
	}

	if !recursive {
		q.Delimiter = "/"
	}

	it := s.gcsClient.Bucket(bucketName).Objects(ctx, q)
	it.PageInfo().MaxSize = int(limit)
	oit := iterator.NewProducer[*storage.ListObjectsResponse_Result]()
	continuationToken := ""

	go func() {
		defer oit.Done()
		count := 0
		for {
			if limit > 0 && count >= int(limit) {
				continuationToken = it.PageInfo().Token
				oit.Error(io.EOF)
				return
			}
			obj, err := it.Next()
			if err == gIterator.Done {
				oit.Error(io.EOF)
				return
			}
			if err != nil {
				oit.Error(err)
				return
			}
			count++

			if obj.Prefix != "" {
				res := &storage.ListObjectsResponse_Result{
					Result: &storage.ListObjectsResponse_Result_Folder{
						Folder: obj.Prefix,
					},
				}
				err := oit.Value(res)
				if err != nil {
					oit.Error(err)
					return
				}
			} else {
				res := &storage.ListObjectsResponse_Result{
					Result: &storage.ListObjectsResponse_Result_Object{
						Object: &storage.Object{
							Path:         obj.Name,
							Size:         uint64(obj.Size),
							Etag:         obj.Etag,
							ContentType:  obj.ContentType,
							LastModified: timestamppb.New(obj.Updated),
						},
					},
				}
				err := oit.Value(res)
				if err != nil {
					oit.Error(err)
					return
				}
			}
		}
	}()

	return continuationToken, oit, nil
}
