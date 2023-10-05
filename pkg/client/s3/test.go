package s3

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Test checks if the s3 client is able to connect.
func (s *Storage) Test(ctx context.Context) error {
	_, err := s.s3.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return err
	}
	return nil
}
