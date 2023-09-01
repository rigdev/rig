package schema

import (
	"github.com/rigdev/rig-go-api/api/v1/service_account"
	"github.com/rigdev/rig/pkg/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/protobuf/proto"
)

type ServiceAccount struct {
	ProjectID        uuid.UUID `bson:"project_id" json:"project_id"`
	ServiceAccountID uuid.UUID `bson:"service_account_id" json:"service_account_id"`
	Name             string    `bson:"name" json:"name"`
	Data             []byte    `bson:"data" json:"data"`
}

func (c ServiceAccount) ToProto() (*service_account.ServiceAccount, error) {
	p := &service_account.ServiceAccount{}
	if err := proto.Unmarshal(c.Data, p); err != nil {
		return nil, err
	}

	return p, nil
}

func ServiceAccountFromProto(projectID, serviceAccountID uuid.UUID, c *service_account.ServiceAccount) (ServiceAccount, error) {
	bs, err := proto.Marshal(c)
	if err != nil {
		return ServiceAccount{}, err
	}

	return ServiceAccount{
		ProjectID:        projectID,
		ServiceAccountID: serviceAccountID,
		Name:             c.GetName(),
		Data:             bs,
	}, nil
}

func GetServiceAccountIDFilter(projectID, serviceAccountID uuid.UUID) bson.M {
	return bson.M{
		"project_id":         projectID,
		"service_account_id": serviceAccountID,
	}
}
