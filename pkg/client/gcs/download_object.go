package gcs

import (
	"context"
	"io"
)

func (s *Storage) DownloadObject(ctx context.Context, bucket, path string) (io.ReadSeekCloser, error) {
	obj := s.gcsClient.Bucket(bucket).Object(path)
	r, err := obj.NewReader(ctx)
	if err != nil {
		return nil, err
	}

	reader := &GSReadSeekCloser{
		r:        r,
		filesize: r.Attrs.Size,
	}
	return reader, nil
}
