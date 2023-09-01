package s3

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/rigdev/rig-go-api/api/v1/storage"
	"github.com/rigdev/rig/pkg/iterator"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Storage) ListBuckets(ctx context.Context) (iterator.Iterator[*storage.Bucket], error) {
	input := s3.ListBucketsInput{}

	output, err := s.s3.ListBuckets(ctx, &input)
	if err != nil {
		return nil, err
	}

	it := iterator.NewProducer[*storage.Bucket]()
	go func() {
		defer it.Done()
		for _, bucket := range output.Buckets {
			b := &storage.Bucket{
				ProviderBucket: *bucket.Name,
				CreatedAt:      timestamppb.New(*bucket.CreationDate),
			}

			locRes, err := s.s3.GetBucketLocation(ctx, &s3.GetBucketLocationInput{
				Bucket: bucket.Name,
			})
			if err != nil {
				it.Error(err)
				return
			}

			b.Region = string(locRes.LocationConstraint)

			if err := it.Value(b); err != nil {
				it.Error(err)
				return
			}
		}
	}()

	return it, nil
}
