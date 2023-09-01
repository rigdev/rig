package storage

import (
	"context"
	"io"

	"github.com/rigdev/rig-go-api/api/v1/storage"
	"github.com/rigdev/rig-go-api/api/v1/storage/settings"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/internal/repository"
	"github.com/rigdev/rig/internal/service/project"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/iterator"
	"github.com/rigdev/rig/pkg/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service struct {
	ps     project.Service
	rs     repository.Storage
	rsec   repository.Secret
	logger *zap.Logger
}

func NewService(logger *zap.Logger, ps project.Service, rs repository.Storage, rsec repository.Secret) *Service {
	return &Service{
		rs:     rs,
		ps:     ps,
		rsec:   rsec,
		logger: logger,
	}
}

func (s *Service) GetBucket(ctx context.Context, bucketName string) (*storage.Bucket, error) {
	_, p, err := s.lookupProviderByBucket(ctx, bucketName)
	if err != nil {
		return nil, err
	}

	for _, b := range p.Buckets {
		if b.Name == bucketName {
			return b, nil
		}
	}
	return nil, errors.NotFoundErrorf("bucket %q not found", bucketName)
}

func (s *Service) GetObject(ctx context.Context, bucketName, path string) (*storage.Object, error) {
	_, p, err := s.lookupProviderByBucket(ctx, bucketName)
	if err != nil {
		return nil, err
	}

	sg, err := s.getStorageGateway(ctx, p)
	if err != nil {
		return nil, err
	}

	var providerBucketName string
	for _, b := range p.Buckets {
		if b.Name == bucketName {
			providerBucketName = b.ProviderBucket
			break
		}
	}

	return sg.GetObject(ctx, providerBucketName, path)
}

func (s *Service) UploadObject(ctx context.Context, reader io.Reader, metadata *storage.UploadObjectRequest_Metadata) (string, uint64, error) {
	_, p, err := s.lookupProviderByBucket(ctx, metadata.GetBucket())
	if err != nil {
		return "", 0, err
	}

	sg, err := s.getStorageGateway(ctx, p)
	if err != nil {
		return "", 0, err
	}

	var providerBucketName string
	for _, b := range p.Buckets {
		if b.Name == metadata.GetBucket() {
			providerBucketName = b.ProviderBucket
			break
		}
	}

	if metadata.GetOnlyCreate() {
		if _, err := s.GetObject(ctx, metadata.Bucket, metadata.Path); errors.IsNotFound(err) {
			// Good, continue.
		} else if err != nil {
			return "", 0, err
		} else {
			return "", 0, errors.FailedPreconditionErrorf("only create is set, but file does exist")
		}
	}

	if metadata.GetOnlyReplace() {
		if _, err := s.GetObject(ctx, metadata.Bucket, metadata.Path); errors.IsNotFound(err) {
			return "", 0, errors.FailedPreconditionErrorf("only replace is set, but file does not exist")
		} else if err != nil {
			return "", 0, err
		}
	}

	return sg.UploadObject(ctx, reader, int64(metadata.GetSize()), providerBucketName, metadata.GetPath(), metadata.GetContentType())
}

func (s *Service) DownloadObject(ctx context.Context, bucketName, path string) (io.ReadSeekCloser, error) {
	_, p, err := s.lookupProviderByBucket(ctx, bucketName)
	if err != nil {
		return nil, err
	}

	sg, err := s.getStorageGateway(ctx, p)
	if err != nil {
		return nil, err
	}

	var providerBucketName string
	for _, b := range p.Buckets {
		if b.Name == bucketName {
			providerBucketName = b.ProviderBucket
			break
		}
	}

	return sg.DownloadObject(ctx, providerBucketName, path)
}

func (s *Service) DeleteObject(ctx context.Context, bucketName, path string) error {
	_, p, err := s.lookupProviderByBucket(ctx, bucketName)
	if err != nil {
		return err
	}

	sg, err := s.getStorageGateway(ctx, p)
	if err != nil {
		return err
	}

	var providerBucketName string
	for _, b := range p.Buckets {
		if b.Name == bucketName {
			providerBucketName = b.ProviderBucket
			break
		}
	}

	return sg.DeleteObject(ctx, providerBucketName, path)
}

func (s *Service) ListObjects(ctx context.Context, bucketName, token, prefix, startpath, endpath string, recursive bool, limit uint32) (string, iterator.Iterator[*storage.ListObjectsResponse_Result], error) {
	_, p, err := s.lookupProviderByBucket(ctx, bucketName)
	if err != nil {
		return "", nil, err
	}

	sg, err := s.getStorageGateway(ctx, p)
	if err != nil {
		return "", nil, err
	}

	var providerBucketName string
	for _, b := range p.Buckets {
		if b.Name == bucketName {
			providerBucketName = b.ProviderBucket
			break
		}
	}

	return sg.ListObjects(ctx, providerBucketName, token, prefix, startpath, endpath, recursive, limit)
}

func (s *Service) CreateBucket(ctx context.Context, name, providerName, region string, providerID uuid.UUID) error {
	provider, err := s.GetProvider(ctx, providerID)
	if err != nil {
		return err
	}

	switch provider.GetConfig().GetConfig().(type) {
	case *storage.Config_Minio:
		provider.GetConfig().GetMinio().Region = region
	case *storage.Config_S3:
		provider.GetConfig().GetS3().Region = region
	}

	sg, err := s.getStorageGateway(ctx, provider)
	if err != nil {
		return err
	}

	pn, err := sg.CreateBucket(ctx, providerName, region)
	if err != nil {
		return err
	}

	bucket := &storage.Bucket{
		Name:           name,
		Region:         region,
		ProviderBucket: pn,
		CreatedAt:      timestamppb.Now(),
	}

	provider.Buckets = append(provider.Buckets, bucket)

	_, err = s.rs.Update(ctx, providerID, provider)
	return err
}

func (s *Service) UnlinkBucket(ctx context.Context, bucket string) error {
	pid, p, err := s.lookupProviderByBucket(ctx, bucket)
	if err != nil {
		return err
	}

	for i, b := range p.Buckets {
		if b.Name == bucket {
			p.Buckets = append(p.Buckets[:i], p.Buckets[i+1:]...)
			break
		}
	}

	_, err = s.rs.Update(ctx, pid, p)
	return err
}

func (s *Service) DeleteBucket(ctx context.Context, bucketName string) error {
	pid, p, err := s.lookupProviderByBucket(ctx, bucketName)
	if err != nil {
		return err
	}

	sg, err := s.getStorageGateway(ctx, p)
	if err != nil {
		return err
	}

	var providerBucketName string
	for i, b := range p.Buckets {
		if b.Name == bucketName {
			providerBucketName = b.ProviderBucket
			p.Buckets = append(p.Buckets[:i], p.Buckets[i+1:]...)
			break
		}
	}

	err = sg.DeleteBucket(ctx, providerBucketName)
	if err != nil {
		return err
	}

	_, err = s.rs.Update(ctx, pid, p)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) ListBuckets(ctx context.Context) (iterator.Iterator[*storage.Bucket], error) {
	pit, _, err := s.ListProviders(ctx, &model.Pagination{})
	if err != nil {
		return nil, err
	}

	bit := iterator.NewProducer[*storage.Bucket]()
	go func() {
		for {
			defer pit.Close()
			defer bit.Done()
			provider, err := pit.Next()
			if err == io.EOF {
				bit.Error(nil)
				return
			} else if err != nil {
				bit.Error(err)
				return
			}
			for _, bucket := range provider.GetBuckets() {
				if err := bit.Value(bucket); err != nil {
					bit.Error(err)
					return
				}
			}
		}
	}()
	return bit, nil
}

func (s *Service) CopyObject(ctx context.Context, dstBucket, dstPath, srcBucket, srcPath string) error {
	_, srcProvider, err := s.lookupProviderByBucket(ctx, srcBucket)
	if err != nil {
		return err
	}

	var srcProviderBucketName string
	for _, b := range srcProvider.Buckets {
		if b.Name == srcBucket {
			srcProviderBucketName = b.ProviderBucket
			break
		}
	}

	_, dstProvider, err := s.lookupProviderByBucket(ctx, dstBucket)
	if err != nil {
		return err
	}

	var dstProviderBucketName string
	for _, b := range dstProvider.Buckets {
		if b.Name == dstBucket {
			dstProviderBucketName = b.ProviderBucket
			break
		}
	}

	if srcProvider.Name == dstProvider.Name {
		sg, err := s.getStorageGateway(ctx, srcProvider)
		if err != nil {
			return err
		}

		return sg.CopyObject(ctx, dstProviderBucketName, dstPath, srcProviderBucketName, srcPath)
	}

	srcSg, err := s.getStorageGateway(ctx, srcProvider)
	if err != nil {
		return err
	}

	dstSg, err := s.getStorageGateway(ctx, dstProvider)
	if err != nil {
		return err
	}

	reader, err := srcSg.DownloadObject(ctx, srcProviderBucketName, srcPath)
	if err != nil {
		return err
	}

	defer reader.Close()

	_, _, err = dstSg.UploadObject(ctx, reader, 0, dstProviderBucketName, dstPath, "")
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) GetSettings(ctx context.Context) (*settings.Settings, error) {
	res := &settings.Settings{}
	err := s.ps.GetSettings(ctx, project.SettingsTypeStorage, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}
