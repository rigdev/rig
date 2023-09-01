package gcs

import (
	"context"
	"strings"
)

func (s *Storage) DeleteObject(ctx context.Context, bucket, path string) error {
	obj := s.gcsClient.Bucket(bucket).Object(strings.TrimPrefix(path, "/"))
	return obj.Delete(ctx)
}
