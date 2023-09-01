package s3

import (
	"context"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func (s *Storage) DownloadObject(ctx context.Context, bucket, path string) (io.ReadSeekCloser, error) {
	path = strings.TrimPrefix(path, "/")

	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(path),
	}

	output, err := s.s3.GetObject(ctx, input)
	if err != nil {
		return nil, err
	}

	// TODO: Fix this to return a ReadSeekCloser
	return struct {
		io.Reader
		io.Seeker
		io.Closer
	}{
		Reader: output.Body,
		Seeker: nil,
		Closer: output.Body,
	}, nil
}
