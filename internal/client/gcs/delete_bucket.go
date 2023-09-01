package gcs

import "context"

func (s *Storage) DeleteBucket(ctx context.Context, name string) error {
	bkt := s.gcsClient.Bucket(name)
	return bkt.Delete(ctx)
}
