package schema

import (
	"reflect"

	"github.com/rigdev/rig-go-api/api/v1/user"
	user_settings "github.com/rigdev/rig-go-api/api/v1/user/settings"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/utils"
	"github.com/rigdev/rig/pkg/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/protobuf/proto"
)

type User struct {
	ProjectID   uuid.UUID `bson:"project_id" json:"project_id"`
	UserID      uuid.UUID `bson:"user_id" json:"user_id"`
	Username    string    `bson:"username,omitempty" json:"username,omitempty"`
	Email       string    `bson:"email,omitempty" json:"email,omitempty"`
	PhoneNumber string    `bson:"phone_number,omitempty" json:"phone_number,omitempty"`
	Search      []string  `bson:"search,omitempty" json:"search,omitempty"`
	Data        []byte    `bson:"data,omitempty" json:"data,omitempty"`
	Password    []byte    `bson:"password,omitempty" json:"password,omitempty"`
}

func (u User) ToProto() (*user.User, error) {
	p := &user.User{}
	if err := proto.Unmarshal(u.Data, p); err != nil {
		return nil, err
	}

	return p, nil
}

func (u User) ToProtoEntry(settings *user_settings.Settings) (*model.UserEntry, error) {
	p, err := u.ToProto()
	if err != nil {
		return nil, err
	}

	e := &model.UserEntry{
		UserId:        u.UserID.String(),
		PrintableName: utils.UserName(p),
		RegisterInfo:  p.GetRegisterInfo(),
		// TODO: check this email verification thing
		Verified:  true,
		CreatedAt: p.GetUserInfo().GetCreatedAt(),
	}

	if settings.GetIsVerifiedEmailRequired() && !p.GetIsEmailVerified() {
		e.Verified = false
	}

	if settings.GetIsVerifiedPhoneRequired() && !p.GetIsPhoneVerified() {
		e.Verified = false
	}

	return e, nil
}

func UserFromProto(projectID uuid.UUID, p *user.User) (User, error) {
	bs, err := proto.Marshal(p)
	if err != nil {
		return User{}, err
	}

	return User{
		ProjectID:   projectID,
		UserID:      uuid.UUID(p.GetUserId()),
		Username:    p.GetUserInfo().GetUsername(),
		Email:       p.GetUserInfo().GetEmail(),
		PhoneNumber: p.GetUserInfo().GetPhoneNumber(),
		Search: []string{
			p.GetUserId(),
			p.GetUserInfo().GetUsername(),
			p.GetUserInfo().GetEmail(),
			p.GetUserInfo().GetPhoneNumber(),
			p.GetProfile().GetFirstName(),
			p.GetProfile().GetLastName(),
		},
		Data: bs,
	}, nil
}

func GetUserIDFilter(projectID, userID uuid.UUID) bson.M {
	return bson.M{
		"project_id": projectID,
		"user_id":    userID,
	}
}

func GetOauth2UserFilter(projectID uuid.UUID, issuer, sub string) bson.M {
	return bson.M{
		"project_id": projectID,
		"iss":        issuer,
		"sub":        sub,
	}
}

func GetUserIdentifierFilter(projectID uuid.UUID, id *model.UserIdentifier) (bson.M, error) {
	switch v := id.GetIdentifier().(type) {
	case *model.UserIdentifier_Email:
		return bson.M{
			"project_id": projectID,
			"email":      v.Email,
		}, nil
	case *model.UserIdentifier_Username:
		return bson.M{
			"project_id": projectID,
			"username":   v.Username,
		}, nil
	default:
		return nil, errors.InvalidArgumentErrorf("invalid user identifier type '%v'", reflect.TypeOf(v))
	}
}
