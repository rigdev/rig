package s3

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go4.org/readerutil"
)

func (s *Storage) UploadObject(ctx context.Context, reader io.Reader, size int64, bucket, path, contentType string) (string, uint64, error) {
	path = strings.TrimPrefix(path, "/")

	fakeSeeker := readerutil.NewFakeSeeker(reader, size)

	input := s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(path),
		Body:          fakeSeeker,
		ContentLength: size,
		ContentType:   aws.String(contentType),
		Metadata: map[string]string{
			"Content-Length": fmt.Sprintf("%d", size),
		},
	}

	_, err := s.s3.PutObject(ctx, &input)
	if err != nil {
		return "", 0, err
	}

	return path, uint64(size), nil
}
