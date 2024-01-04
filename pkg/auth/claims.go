package auth

import (
	"context"

	"github.com/golang-jwt/jwt"
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

type State struct {
	Rand        string
	ProjectID   string
	AppRedirect string
}
