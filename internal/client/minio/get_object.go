package minio

import (
	"context"

	"github.com/minio/minio-go/v7"
	"github.com/rigdev/rig-go-api/api/v1/storage"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Storage) GetObject(ctx context.Context, bucketName, path string) (*storage.Object, error) {
	objInfo, err := s.minioClient.StatObject(ctx, bucketName, path, minio.GetObjectOptions{})
	if err != nil {
		return nil, toError(err)
	}

	return &storage.Object{
		Path:         path,
		LastModified: timestamppb.New(objInfo.LastModified),
		Size:         uint64(objInfo.Size),
		Etag:         objInfo.ETag,
		ContentType:  objInfo.ContentType,
	}, nil
}
