package auth

import (
	"context"

	"github.com/golang-jwt/jwt"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
)

type claimsIDKeyType string

const _claimsIDKey claimsIDKeyType = "claimsID"

type SubjectType int

const (
	SubjectTypeInvalid SubjectType = iota
	SubjectTypeUser
	SubjectTypeServiceAccount
)

func WithClaims(ctx context.Context, c Claims) context.Context {
	return context.WithValue(ctx, _claimsIDKey, c)
}

func GetClaims(ctx context.Context) (Claims, error) {
	val, ok := ctx.Value(_claimsIDKey).(Claims)
	if ok {
		return val, nil
	}

	return nil, errors.UnauthenticatedErrorf("unauthenticated request")
}

type Claims interface {
	jwt.Claims

	GetIssuer() string
	GetProjectID() string
	GetSubject() uuid.UUID
	GetSubjectType() SubjectType
	GetSessionID() uuid.UUID
}

type Oauth2Claims struct {
	Sub       string `json:"sub"`
	Iss       string `json:"iss"`
	Image     string `json:"image"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// type OAuthProvider struct {
// 	ClientID     string              `bson:"client_id,omitempty" json:"client_id,omitempty"`
// 	ClientSecret string              `bson:"client_secret,omitempty" json:"client_secret,omitempty"`
// 	Enabled      bool                `bson:"enabled,omitempty" json:"enabled,omitempty"`
// 	Url          string              `bson:"url,omitempty" json:"url,omitempty"`
// 	Type         model.OauthProvider `bson:"type,omitempty" json:"type,omitempty"`
// }

// type OAuthProviders struct {
// 	RedirectURL  string         `bson:"redirect_url,omitempty" json:"redirect_url,omitempty"`
// 	CallbackURLs []string       `bson:"callback_urls,omitempty" json:"callback_urls,omitempty"`
// 	Google       *OAuthProvider `bson:"google,omitempty" json:"google,omitempty"`
// 	Github       *OAuthProvider `bson:"github,omitempty" json:"github,omitempty"`
// 	Facebook     *OAuthProvider `bson:"facebook,omitempty" json:"facebook,omitempty"`
// }

// func (provs *OAuthProviders) ToProto() *settings.OauthSettings {
// 	if provs == nil {
// 		return nil
// 	}
// 	return &settings.OauthSettings{
// 		RigCallbackUrl: provs.RedirectURL,
// 		CallbackUrls:      provs.CallbackURLs,
// 		Google:            provs.Google.ToProto(),
// 		Github:            provs.Github.ToProto(),
// 		Facebook:          provs.Facebook.ToProto(),
// 	}
// }

// func (prov *OAuthProvider) ToProto() *settings.OauthProviderSettings {
// 	if prov == nil {
// 		return nil
// 	}
// 	return &settings.OauthProviderSettings{
// 		ClientId:      prov.ClientID,
// 		ClientSecret:  prov.ClientSecret,
// 		Provider:      prov.Type,
// 		ProviderUrl:   prov.Url,
// 		AllowLogin:    prov.Enabled,
// 		AllowRegister: prov.Enabled,
// 	}
// }

type State struct {
	Rand         string
	ProjectID    string
	AppRedirect  string
	ProviderType model.OauthProvider
}
