package gcs

import (
	"context"
	"io"
	"strings"
)

func (s *Storage) CopyObject(ctx context.Context, dstBucket, dstPath, srcBucket, srcPath string) error {
	srcObj := s.gcsClient.Bucket(srcBucket).Object(strings.TrimPrefix(srcPath, "/"))
	r, err := srcObj.NewReader(ctx)
	if err != nil {
		return err
	}

	dstObj := s.gcsClient.Bucket(dstBucket).Object(strings.TrimPrefix(dstPath, "/"))
	w := dstObj.NewWriter(ctx)

	_, err = io.Copy(w, r)
	if err != nil {
		return err
	}

	if err := w.Close(); err != nil {
		return err
	}
	return nil

}
