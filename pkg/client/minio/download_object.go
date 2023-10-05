package minio

import (
	"context"
	"io"

	"github.com/minio/minio-go/v7"
)

func (s *Storage) DownloadObject(ctx context.Context, bucketName, path string) (io.ReadSeekCloser, error) {
	r, err := s.minioClient.GetObject(ctx, bucketName, path, minio.GetObjectOptions{})
	if err != nil {
		return nil, toError(err)
	}

	if _, err := r.Stat(); err != nil {
		return nil, toError(err)
	}
	return r, nil
}
