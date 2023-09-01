package auth

import (
	"github.com/golang-jwt/jwt"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/uuid"
)

type RigClaims struct {
	ProjectID   uuid.UUID        `json:"pid"`
	Subject     uuid.UUID        `json:"sub"`
	SubjectType auth.SubjectType `json:"sty"`
	SessionID   uuid.UUID        `json:"jti,omitempty"`

	jwt.StandardClaims
}

func (c RigClaims) GetIssuer() string                { return c.Issuer }
func (c RigClaims) GetProjectID() uuid.UUID          { return c.ProjectID }
func (c RigClaims) GetSubject() uuid.UUID            { return c.Subject }
func (c RigClaims) GetSubjectType() auth.SubjectType { return c.SubjectType }
func (c RigClaims) GetSessionID() uuid.UUID          { return c.SessionID }

// The metadata stored in all access tokens.
type AccessClaims struct {
	Groups   []uuid.UUID            `json:"gps,omitempty"`
	MetaData map[string]interface{} `json:"meta_data,omitempty"`

	RigClaims
}

// The metadata stored in all refresh tokens.
type RefreshClaims struct {
	Groups []uuid.UUID `json:"gps,omitempty"`

	RigClaims
}

type ProjectClaims struct {
	UseProjectID uuid.UUID `json:"aid"`

	RigClaims
}
