package schema

import (
	"github.com/rigdev/rig/gen/go/oauth2"
	"github.com/rigdev/rig/pkg/uuid"
	"google.golang.org/protobuf/proto"
)

type Oauth2Link struct {
	ProjectID uuid.UUID `bson:"project_id" json:"project_id"`
	UserID    uuid.UUID `bson:"user_id" json:"user_id"`
	Issuer    string    `bson:"iss" json:"iss"`
	Subject   string    `bson:"sub" json:"sub"`
	Data      []byte    `bson:"data,omitempty" json:"data,omitempty"`
}

func (u Oauth2Link) ToProto() (*oauth2.ProviderLink, error) {
	p := &oauth2.ProviderLink{}
	if err := proto.Unmarshal(u.Data, p); err != nil {
		return nil, err
	}

	return p, nil
}

func Oauth2LinkFromProto(projectID, userID uuid.UUID, p *oauth2.ProviderLink) (Oauth2Link, error) {
	bs, err := proto.Marshal(p)
	if err != nil {
		return Oauth2Link{}, err
	}

	return Oauth2Link{
		ProjectID: projectID,
		UserID:    userID,
		Issuer:    p.GetIssuer(),
		Subject:   p.GetSubject(),
		Data:      bs,
	}, nil
}
