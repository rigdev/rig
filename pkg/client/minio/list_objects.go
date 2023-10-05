package minio

import (
	"context"
	"io"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/rigdev/rig-go-api/api/v1/storage"
	"github.com/rigdev/rig/pkg/iterator"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Storage) ListObjects(ctx context.Context, bucketName, token, prefix,
	startpath, endpath string, recursive bool, limit uint32,
) (string, iterator.Iterator[*storage.ListObjectsResponse_Result], error) {
	prefix = strings.TrimPrefix(prefix, "/")

	if token != "" {
		startpath = token
	}

	objectCh := s.minioClient.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
		Prefix:       prefix,
		Recursive:    recursive,
		MaxKeys:      int(limit),
		StartAfter:   startpath,
		WithMetadata: true,
	})
	it := iterator.NewProducer[*storage.ListObjectsResponse_Result]()

	newToken := ""

	go func() {
		var count int
		defer it.Done()
		for object := range objectCh {
			count++
			if limit > 0 && count == int(limit) {
				newToken = object.Key
			}
			if object.Err != nil {
				it.Error(object.Err)
				return
			}
			resp, err := objToNunt(object, endpath, prefix)
			if err != nil {
				it.Error(err)
				return
			}
			if err := it.Value(resp); err != nil {
				it.Error(err)
				return
			}
		}
	}()
	return newToken, it, nil
}

func objToNunt(obj minio.ObjectInfo, endpath, prefix string) (*storage.ListObjectsResponse_Result, error) {
	if obj.Err != nil {
		return nil, obj.Err
	}

	if endpath != "" && strings.Compare(obj.Key, endpath) >= 0 {
		return nil, io.EOF
	}

	key := "/" + obj.Key

	if strings.HasSuffix(key, "/") {
		return &storage.ListObjectsResponse_Result{
			Result: &storage.ListObjectsResponse_Result_Folder{
				Folder: key,
			},
		}, nil
	}

	contentType := obj.UserMetadata["content-type"]

	return &storage.ListObjectsResponse_Result{
		Result: &storage.ListObjectsResponse_Result_Object{
			Object: &storage.Object{
				Path:         key,
				LastModified: timestamppb.New(obj.LastModified),
				Size:         uint64(obj.Size),
				Etag:         obj.ETag,
				ContentType:  contentType,
			},
		},
	}, nil
}
