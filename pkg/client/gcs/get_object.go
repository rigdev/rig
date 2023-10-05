package gcs

import (
	"context"
	"strings"

	"github.com/rigdev/rig-go-api/api/v1/storage"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Storage) GetObject(ctx context.Context, bucketName, path string) (*storage.Object, error) {
	objectInfo := s.gcsClient.Bucket(bucketName).Object(strings.TrimPrefix(path, "/"))

	attrs, err := objectInfo.Attrs(ctx)
	if err != nil {
		return nil, err
	}

	return &storage.Object{
		Path:         path,
		LastModified: timestamppb.New(attrs.Updated),
		Size:         uint64(attrs.Size),
		Etag:         attrs.Etag,
		ContentType:  attrs.ContentType,
	}, nil
}
