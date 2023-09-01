package gcs

import (
	"context"

	"cloud.google.com/go/storage"
)

func (s *Storage) ComposeObject(ctx context.Context, bucketName string, dest string, srcs ...string) error {
	bkt := s.gcsClient.Bucket(bucketName)

	srcsHandler := make([]*storage.ObjectHandle, len(srcs))
	for i, src := range srcs {
		srcsHandler[i] = bkt.Object(src)
	}
	_, err := bkt.Object(dest).ComposerFrom(srcsHandler...).Run(ctx)
	if err != nil {
		return err
	}

	return nil
}
