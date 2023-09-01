package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"reflect"

	"github.com/rigdev/rig-go-api/api/v1/storage"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/internal/client/gcs"
	"github.com/rigdev/rig/internal/client/minio"
	"github.com/rigdev/rig/internal/client/s3"
	storage_gateway "github.com/rigdev/rig/internal/gateway/storage"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/iterator"
	"github.com/rigdev/rig/pkg/uuid"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Service) LookupProvider(ctx context.Context, name string) (uuid.UUID, *storage.Provider, error) {
	pid, p, sid, err := s.rs.Lookup(ctx, name)
	if err != nil {
		return uuid.Nil, nil, err
	}

	err = s.getProviderSecret(ctx, sid, p)
	if err != nil {
		return pid, p, err
	}

	return pid, p, err
}

func (s *Service) CreateProvider(ctx context.Context, name string, config *storage.Config, linkBuckets bool) (uuid.UUID, *storage.Provider, error) {
	providerID := uuid.New()

	p := &storage.Provider{
		Name:      name,
		Config:    config,
		CreatedAt: timestamppb.Now(),
	}

	sg, err := s.getStorageGateway(ctx, p)
	if err != nil {
		return uuid.Nil, nil, err
	}

	err = sg.Test(ctx)
	if err != nil {
		return uuid.Nil, nil, err
	}

	if linkBuckets {
		bucketsIt, err := sg.ListBuckets(ctx)
		if err != nil {
			return uuid.Nil, nil, err
		}
		defer bucketsIt.Close()

		rigBucketsIt, err := s.ListBuckets(ctx)
		if err != nil {
			return uuid.Nil, nil, err
		}
		defer rigBucketsIt.Close()

		rigBuckets, err := iterator.Collect(rigBucketsIt)
		if err != nil {
			return uuid.Nil, nil, err
		}

		buckets := []*storage.Bucket{}
		for {
			b, err := bucketsIt.Next()
			if err == io.EOF {
				p.Buckets = buckets
				break
			} else if err != nil {
				return uuid.Nil, nil, err
			}

			if !slices.ContainsFunc(rigBuckets, func(bb *storage.Bucket) bool {
				return b.ProviderBucket == bb.Name
			}) {
				b.Name = b.ProviderBucket
				buckets = append(buckets, b)
			} else {
				ptype, err := GetProviderType(p.GetConfig())
				if err != nil {
					fmt.Println("Skipping bucket")
					continue
				}
				b.Name = ptype + "-" + b.ProviderBucket
				buckets = append(buckets, b)
			}
		}
	}

	sID := uuid.New()
	var secret []byte

	switch config.Config.(type) {
	case *storage.Config_Gcs:
		secret = config.GetGcs().GetConfig()
		config.GetGcs().Config = nil
	case *storage.Config_S3:
		secret, err = json.Marshal(config.GetS3().GetCredentials())
		if err != nil {
			return uuid.Nil, nil, err
		}

		config.GetS3().Credentials = nil
	case *storage.Config_Minio:
		secret, err = json.Marshal(config.GetMinio().GetCredentials())
		if err != nil {
			return uuid.Nil, nil, err
		}

		config.GetMinio().Credentials = nil
	}

	err = s.rsec.Create(ctx, sID, secret)
	if err != nil {
		return uuid.Nil, nil, err
	}

	p, err = s.rs.Create(ctx, providerID, sID, p)
	if err != nil {
		return uuid.Nil, nil, err
	}

	return providerID, p, nil
}

func (s *Service) DeleteProvider(ctx context.Context, providerID uuid.UUID) error {
	_, sid, err := s.rs.Get(ctx, providerID)
	if err != nil {
		return err
	}

	err = s.rsec.Delete(ctx, sid)
	if err != nil {
		return err
	}

	return s.rs.Delete(ctx, providerID)
}

func (s *Service) ListProviders(ctx context.Context, pagination *model.Pagination) (iterator.Iterator[*storage.ProviderEntry], uint64, error) {
	return s.rs.List(ctx, pagination)
}

func (s *Service) GetProvider(ctx context.Context, providerID uuid.UUID) (*storage.Provider, error) {
	p, sid, err := s.rs.Get(ctx, providerID)
	if err != nil {
		return nil, err
	}

	err = s.getProviderSecret(ctx, sid, p)
	if err != nil {
		return nil, err
	}

	return p, err
}

func (s *Service) getStorageGateway(ctx context.Context, provider *storage.Provider) (storage_gateway.Gateway, error) {
	switch v := provider.GetConfig().GetConfig().(type) {
	case *storage.Config_Gcs:
		return gcs.New(ctx, v.Gcs)
	case *storage.Config_S3:
		return s3.New(v.S3)
	case *storage.Config_Minio:
		return minio.New(v.Minio)
	default:
		return nil, errors.InvalidArgumentErrorf("invalid storage provider type '%v'", reflect.TypeOf(v))
	}
}

func (s *Service) lookupProviderByBucket(ctx context.Context, bucketName string) (uuid.UUID, *storage.Provider, error) {
	id, provider, sid, err := s.rs.LookupByBucket(ctx, bucketName)
	if err != nil {
		return uuid.Nil, nil, err
	}
	switch provider.GetConfig().GetConfig().(type) {
	case *storage.Config_S3:
		var region string
		for _, b := range provider.Buckets {
			if b.Name == bucketName {
				region = b.Region
				provider.Config.GetS3().Region = region
				break
			}
		}
	case *storage.Config_Minio:
		var region string
		for _, b := range provider.Buckets {
			if b.Name == bucketName {
				region = b.Region
				provider.Config.GetMinio().Region = region
				break
			}
		}
	}

	err = s.getProviderSecret(ctx, sid, provider)
	if err != nil {
		return uuid.Nil, nil, err
	}

	return id, provider, nil
}

func GetProviderType(p *storage.Config) (string, error) {
	switch p.GetConfig().(type) {
	case *storage.Config_S3:
		return "s3", nil
	case *storage.Config_Gcs:
		return "gcs", nil
	case *storage.Config_Minio:
		return "minio", nil
	default:
		return "", errors.InvalidArgumentErrorf("unknown provider type")
	}
}

func (s *Service) getProviderSecret(ctx context.Context, sid uuid.UUID, p *storage.Provider) error {
	secret, err := s.rsec.Get(ctx, sid)
	if err != nil {
		return err
	}

	switch p.GetConfig().GetConfig().(type) {
	case *storage.Config_Gcs:
		p.GetConfig().GetGcs().Config = secret
	case *storage.Config_S3:
		p.GetConfig().GetS3().Credentials = &model.ProviderCredentials{}
		err = json.Unmarshal(secret, p.GetConfig().GetS3().Credentials)
		if err != nil {
			return err
		}
	case *storage.Config_Minio:
		p.GetConfig().GetMinio().Credentials = &model.ProviderCredentials{}
		err = json.Unmarshal(secret, p.GetConfig().GetMinio().GetCredentials())
		if err != nil {
			return err
		}
	default:
		return errors.InvalidArgumentErrorf("invalid storage provider type")
	}

	return nil
}
