package minio

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/storage"
	"github.com/rigdev/rig/pkg/iterator"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Storage) ListBuckets(ctx context.Context) (iterator.Iterator[*storage.Bucket], error) {
	minioBuckets, err := s.minioClient.ListBuckets(ctx)
	if err != nil {
		return nil, err
	}

	it := iterator.NewProducer[*storage.Bucket]()
	go func() {
		defer it.Done()
		for _, bucket := range minioBuckets {
			rigBucket := &storage.Bucket{
				ProviderBucket: bucket.Name,
				CreatedAt:      timestamppb.New(bucket.CreationDate),
			}

			region, err := s.minioClient.GetBucketLocation(ctx, bucket.Name)
			if err != nil {
				it.Error(err)
				return
			}

			rigBucket.Region = region

			if err := it.Value(rigBucket); err != nil {
				it.Error(err)
				return
			}
		}
	}()
	return it, nil
}
