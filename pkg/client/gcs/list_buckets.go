package gcs

import (
	"context"
	"io"

	"github.com/rigdev/rig-go-api/api/v1/storage"
	"github.com/rigdev/rig/pkg/iterator"
	gIterator "google.golang.org/api/iterator"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Storage) ListBuckets(ctx context.Context) (iterator.Iterator[*storage.Bucket], error) {
	it := s.gcsClient.Buckets(ctx, s.projectID)
	bit := iterator.NewProducer[*storage.Bucket]()
	go func() {
		defer bit.Done()
		for {
			bktAttrs, err := it.Next()
			if err == gIterator.Done {
				bit.Error(io.EOF)
				return
			}
			if err != nil {
				bit.Error(err)
			}
			b := &storage.Bucket{
				Region:         bktAttrs.Location,
				ProviderBucket: bktAttrs.Name,
				CreatedAt:      timestamppb.New(bktAttrs.Created),
			}
			if err := bit.Value(b); err != nil {
				bit.Error(err)
				return
			}
		}
	}()
	return bit, nil
}
