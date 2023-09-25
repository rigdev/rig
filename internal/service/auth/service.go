package auth

import (
	"context"
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/rigdev/rig-go-api/api/v1/authentication"
	api_user "github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig-go-api/api/v1/user/settings"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/internal/config"
	"github.com/rigdev/rig/internal/oauth2"
	"github.com/rigdev/rig/internal/repository"
	"github.com/rigdev/rig/internal/service/project"
	"github.com/rigdev/rig/internal/service/user"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/utils"
	"github.com/rigdev/rig/pkg/uuid"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service struct {
	cfg config.Config
	sr  repository.Session
	rsa repository.ServiceAccount
	rs  repository.Secret
	us  user.Service
	ps  project.Service
	// cert       *x509.Certificate
	issuer     string
	certBytes  []byte
	publicKey  interface{}
	privateKey crypto.PrivateKey
	sm         jwt.SigningMethod
	vcr        repository.VerificationCode
	oauth2     *oauth2.Providers
	logger     *zap.Logger
}

type newServiceParams struct {
	fx.In

	Config             config.Config
	SessionRepo        repository.Session
	ServiceAccountRepo repository.ServiceAccount
	SecretRepo         repository.Secret
	UserService        user.Service
	ProjService        project.Service
	Vcr                repository.VerificationCode
	Oauth2             *oauth2.Providers
	Logger             *zap.Logger
}

func NewService(p newServiceParams) (*Service, error) {
	s := &Service{
		cfg:    p.Config,
		sr:     p.SessionRepo,
		rsa:    p.ServiceAccountRepo,
		rs:     p.SecretRepo,
		us:     p.UserService,
		ps:     p.ProjService,
		vcr:    p.Vcr,
		oauth2: p.Oauth2,
		logger: p.Logger,
	}

	switch {
	case
		p.Config.Auth.JWT.CertificateFile != "",
		p.Config.Auth.JWT.CertificateKeyFile != "":
		cert, err := tls.LoadX509KeyPair(p.Config.Auth.JWT.CertificateFile, p.Config.Auth.JWT.CertificateKeyFile)
		if err != nil {
			return nil, err
		}

		x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
		if err != nil {
			return nil, err
		}

		switch x509Cert.PublicKeyAlgorithm {
		case x509.RSA:
			s.sm = jwt.SigningMethodRS256
		case x509.Ed25519:
			s.sm = jwt.SigningMethodEdDSA
		case x509.ECDSA:
			s.sm = jwt.SigningMethodES256
		default:
			return nil, fmt.Errorf("unsupported certificate algorithm")
		}

		s.issuer = x509Cert.Issuer.CommonName
		s.publicKey = x509Cert.PublicKey
		s.certBytes = cert.Certificate[0]
		s.privateKey = cert.PrivateKey

	case p.Config.Auth.JWT.Secret != "":
		s.sm = jwt.SigningMethodHS512
		s.publicKey = []byte(p.Config.Auth.JWT.Secret)
		s.privateKey = []byte(p.Config.Auth.JWT.Secret)

	default:
		return nil, fmt.Errorf("missing JWT configuration")
	}

	return s, nil
}

func (s *Service) GetJWTMethod() (jwt.SigningMethod, string) {
	switch s.sm.(type) {
	case *jwt.SigningMethodHMAC:
		return s.sm, string(s.publicKey.([]byte))

	default:
		bs := pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: s.certBytes,
		})
		return s.sm, string(bs)
	}
}

func (s *Service) RefreshToken(ctx context.Context, oldRefreshToken string) (*authentication.Token, error) {
	c, err := s.validateRefreshToken(ctx, oldRefreshToken)
	if err != nil {
		return nil, err
	}

	// We're validating the current user, switch to token project ID.
	ctx = auth.WithProjectID(ctx, c.ProjectID)

	switch c.GetSubjectType() {
	case auth.SubjectTypeUser:
		sess, err := s.sr.Get(ctx, c.GetSubject(), c.SessionID)
		if err != nil {
			return nil, err
		}

		if sess.GetIsInvalidated() {
			return nil, errors.PermissionDeniedErrorf("session invalidated")
		}
		u, err := s.us.GetUser(ctx, c.GetSubject())
		if err != nil {
			return nil, err
		}

		if time.Unix(c.IssuedAt, 0).Before(u.GetNewSessionsSince().AsTime()) {
			return nil, errors.PermissionDeniedErrorf("session expired")
		}

	case auth.SubjectTypeServiceAccount:
		pID, _, err := s.rsa.Get(ctx, c.GetSubject())
		if err != nil {
			return nil, err
		}

		if pID != c.GetProjectID() {
			return nil, errors.PermissionDeniedErrorf("invalid refresh token")
		}

	default:
		return nil, errors.PermissionDeniedErrorf("invalid refresh token")
	}

	set, err := s.us.GetSettings(ctx)
	if err != nil {
		return nil, err
	}

	var ss *api_user.Session
	if c.GetSubjectType() == auth.SubjectTypeUser {
		if ss, err = s.sr.Get(ctx, c.GetSubject(), c.SessionID); err != nil {
			return nil, err
		}
	}

	return s.generateToken(ctx, c.SessionID, ss, c.GetProjectID(), c.GetSubject(), c.GetSubjectType(), nil, set)
}

func (s *Service) UseProject(ctx context.Context, projectID string) (string, error) {
	s.logger.Debug("authenticating for project", zap.String("project_id", projectID))
	ctx = auth.WithProjectID(ctx, projectID)
	_, err := s.ps.GetProject(ctx)
	if err != nil {
		return "", err
	}

	c, err := auth.GetClaims(ctx)
	if err != nil {
		return "", err
	}

	pc := &ProjectClaims{
		RigClaims: RigClaims{
			ProjectID:   c.GetProjectID(),
			Subject:     c.GetSubject(),
			SubjectType: c.GetSubjectType(),
		},

		UseProjectID: projectID,
	}

	pc.Issuer = s.issuer
	pc.ExpiresAt = time.Now().Add(1 * time.Hour).Unix()
	pc.IssuedAt = time.Now().Unix()

	t, err := s.generateTokenClaims(ctx, pc)
	if err != nil {
		return "", err
	}

	return t, nil
}

func (s *Service) ValidateAccessToken(ctx context.Context, jwtToken string) (*AccessClaims, error) {
	t := &AccessClaims{}
	return t, s.validateToken(ctx, jwtToken, t)
}

func (s *Service) validateRefreshToken(ctx context.Context, jwtToken string) (*RefreshClaims, error) {
	t := &RefreshClaims{}

	if err := s.validateToken(ctx, jwtToken, t); err != nil {
		return nil, err
	}

	return t, nil
}

func (s *Service) ValidateProjectToken(ctx context.Context, jwtToken string) (*ProjectClaims, error) {
	t := &ProjectClaims{}
	return t, s.validateToken(ctx, jwtToken, t)
}

func (s *Service) generateToken(ctx context.Context, sessionID uuid.UUID, ss *api_user.Session, projectID string, subject uuid.UUID, subjectType auth.SubjectType, gs []uuid.UUID, set *settings.Settings) (*authentication.Token, error) {
	expiresAt := time.Now().Add(set.GetRefreshTokenTtl().AsDuration())
	refreshToken, err := s.generateRefreshToken(ctx, sessionID, projectID, subject, subjectType, gs, expiresAt)
	if err != nil {
		return nil, err
	}

	accessToken, err := s.generateAccessToken(ctx, sessionID, projectID, subject, subjectType, gs, time.Now().Add(set.GetAccessTokenTtl().AsDuration()))
	if err != nil {
		return nil, err
	}

	if ss != nil {
		ns := proto.Clone(ss).(*api_user.Session)
		ns.RenewedAt = timestamppb.Now()
		ns.ExpiresAt = timestamppb.New(expiresAt)
		if err := s.sr.Update(ctx, subject, sessionID, ns); err != nil {
			return nil, err
		}
	}

	return &authentication.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *Service) validateToken(ctx context.Context, jwtToken string, c auth.Claims) error {
	token, err := jwt.ParseWithClaims(
		jwtToken,
		c,
		func(token *jwt.Token) (interface{}, error) {
			return s.publicKey, nil
		},
	)
	if err != nil {
		return err
	}
	if !token.Valid {
		return errors.PermissionDeniedErrorf("invalid token")
	}

	if c.GetIssuer() != s.issuer {
		return errors.PermissionDeniedErrorf("invalid token issuer")
	}

	return nil
}

func (s *Service) generateAccessToken(ctx context.Context, sessionID uuid.UUID, projectID string, subject uuid.UUID, subjectType auth.SubjectType, groups []uuid.UUID, expiresAt time.Time) (string, error) {
	claims := &AccessClaims{
		RigClaims: RigClaims{
			ProjectID:   projectID,
			Subject:     subject,
			SubjectType: subjectType,
			SessionID:   sessionID,
		},
		Groups: groups,
	}
	claims.Issuer = s.issuer
	claims.ExpiresAt = expiresAt.Unix()
	claims.IssuedAt = time.Now().Unix()

	return s.generateTokenClaims(ctx, claims)
}

func (s *Service) generateRefreshToken(ctx context.Context, sessionID uuid.UUID, projectID string, subject uuid.UUID, subjectType auth.SubjectType, groups []uuid.UUID, expiresAt time.Time) (string, error) {
	claims := &RefreshClaims{
		RigClaims: RigClaims{
			ProjectID:   projectID,
			Subject:     subject,
			SubjectType: subjectType,
			SessionID:   sessionID,
		},

		Groups: groups,
	}

	claims.Issuer = s.issuer
	claims.ExpiresAt = expiresAt.Unix()
	claims.IssuedAt = time.Now().Unix()

	return s.generateTokenClaims(ctx, claims)
}

func (s *Service) generateTokenClaims(ctx context.Context, c auth.Claims) (string, error) {
	if c.GetProjectID() == "" {
		return "", fmt.Errorf("invalid token subject")
	}

	if c.GetSubject().IsNil() {
		return "", fmt.Errorf("invalid token subject")
	}

	signedToken, err := jwt.NewWithClaims(s.sm, c).SignedString(s.privateKey)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func (s *Service) GetAuthor(ctx context.Context) (*model.Author, error) {
	a := &model.Author{}
	ac, err := auth.GetClaims(ctx)
	if errors.IsUnauthenticated(err) {
		if _, err := auth.GetProjectID(ctx); err != nil {
			return nil, err
		}
		a.PrintableName = "system"
		return a, nil
	}
	if err != nil {
		return nil, err
	}

	switch ac.GetSubjectType() {
	case auth.SubjectTypeUser:
		a.Account = &model.Author_UserId{
			UserId: ac.GetSubject().String(),
		}

		if u, err := s.us.GetUser(auth.WithProjectID(ctx, auth.RigProjectID), ac.GetSubject()); errors.IsNotFound(err) {
		} else if err != nil {
			return nil, err
		} else {
			a.Identifier = utils.UserIdentifier(u)
			a.PrintableName = utils.UserName(u)
		}

	case auth.SubjectTypeServiceAccount:
		a.Account = &model.Author_ServiceAccountId{
			ServiceAccountId: ac.GetSubject().String(),
		}

		if _, c, err := s.rsa.Get(ctx, ac.GetSubject()); errors.IsNotFound(err) {
		} else if err != nil {
			return nil, err
		} else {
			a.Identifier = c.GetName()
			a.PrintableName = c.GetName()
		}
	default:
		return nil, errors.InvalidArgumentErrorf("unknown subject type '%v'", ac.GetSubjectType())
	}

	return a, nil
}
