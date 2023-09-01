package schema

import (
	"github.com/rigdev/rig-go-api/api/v1/storage"
	"github.com/rigdev/rig/pkg/uuid"
	"google.golang.org/protobuf/proto"
)

type Provider struct {
	ProjectID  uuid.UUID         `bson:"project_id" json:"project_id"`
	ProviderID uuid.UUID         `bson:"provider_id" json:"provider_id"`
	SecretID   uuid.UUID         `bson:"secrets_id" json:"secrets_id"`
	Name       string            `bson:"name" json:"name"`
	Buckets    []*storage.Bucket `bson:"buckets" json:"buckets"`
	Data       []byte            `bson:"data" json:"data"`
}

func (p *Provider) ToProto() (*storage.Provider, error) {
	pr := &storage.Provider{}
	if err := proto.Unmarshal(p.Data, pr); err != nil {
		return nil, err
	}

	pr.Buckets = p.Buckets

	return pr, nil
}

func (p *Provider) ToProtoEntry() (*storage.ProviderEntry, error) {
	pp, err := p.ToProto()
	if err != nil {
		return nil, err
	}

	e := &storage.ProviderEntry{
		ProviderId: p.ProviderID.String(),
		Name:       p.Name,
		Config:     pp.GetConfig(),
		Buckets:    p.Buckets,
		CreatedAt:  pp.CreatedAt,
	}

	return e, nil
}

func ProviderFromProto(projectID, providerID, secretsID uuid.UUID, p *storage.Provider) (*Provider, error) {
	bs, err := proto.Marshal(p)
	if err != nil {
		return nil, err
	}

	return &Provider{
		ProjectID:  projectID,
		ProviderID: providerID,
		SecretID:   secretsID,
		Name:       p.GetName(),
		Buckets:    p.GetBuckets(),
		Data:       bs,
	}, nil
}
