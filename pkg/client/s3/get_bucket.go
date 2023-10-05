package s3

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/rigdev/rig-go-api/api/v1/storage"
	"github.com/rigdev/rig/pkg/errors"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Storage) GetBucket(ctx context.Context, name string) (*storage.Bucket, error) {
	input := s3.ListBucketsInput{}

	output, err := s.s3.ListBuckets(ctx, &input)
	if err != nil {
		return nil, err
	}

	var bucket *storage.Bucket
	for _, b := range output.Buckets {
		if *b.Name == name {
			bucket = &storage.Bucket{
				ProviderBucket: *b.Name,
				CreatedAt:      timestamppb.New(*b.CreationDate),
			}
			break
		}
	}
	if bucket == nil {
		return nil, errors.NotFoundErrorf("bucket %s not found", name)
	}

	locRes, err := s.s3.GetBucketLocation(ctx, &s3.GetBucketLocationInput{
		Bucket: aws.String(name),
	})
	if err != nil {
		return nil, err
	}

	bucket.Region = string(locRes.LocationConstraint)

	return bucket, nil
}
