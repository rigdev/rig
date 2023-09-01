package gcs

import (
	"context"
	"io"
	"strings"
)

func (s *Storage) UploadObject(ctx context.Context, reader io.Reader, size int64, bucket, path, contentType string) (string, uint64, error) {
	obj := s.gcsClient.Bucket(bucket).Object(strings.TrimPrefix(path, "/"))
	w := obj.NewWriter(ctx)
	w.Size = size

	w.ContentType = contentType

	written, err := io.Copy(w, reader)
	if err != nil {
		return "", 0, err
	}

	if err := w.Close(); err != nil {
		return "", 0, err
	}

	return path, uint64(written), nil
}
