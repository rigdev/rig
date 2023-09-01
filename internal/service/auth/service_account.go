package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/rigdev/rig-go-api/api/v1/authentication"
	"github.com/rigdev/rig-go-api/api/v1/service_account"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/hash"
	"github.com/rigdev/rig/pkg/iterator"
	"github.com/rigdev/rig/pkg/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Service) CreateServiceAccount(ctx context.Context, name string, managed bool) (*service_account.ServiceAccount, string, string, error) {
	us, err := s.us.GetSettings(ctx)
	if err != nil {
		return nil, "", "", err
	}

	if name == "" {
		return nil, "", "", errors.InvalidArgumentErrorf("service account name cannot be empty")
	}

	serviceAccountID := uuid.New()
	sa := &service_account.ServiceAccount{
		Name:      name,
		CreatedAt: timestamppb.Now(),
	}
	sa.CreatedBy, err = s.GetAuthor(ctx)
	if err != nil {
		return nil, "", "", err
	}

	if err := s.rsa.Create(ctx, serviceAccountID, sa); err != nil {
		return nil, "", "", err
	}

	clientSecretRaw := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, clientSecretRaw); err != nil {
		return nil, "", "", err
	}

	clientSecret := fmt.Sprint("secret_", hex.EncodeToString(clientSecretRaw))

	h := hash.New(us.GetPasswordHashing())
	pw, err := h.Generate(clientSecret)
	if err != nil {
		return nil, "", "", err
	}

	if err := s.rsa.UpdateClientSecret(ctx, serviceAccountID, pw); err != nil {
		return nil, "", "", err
	}

	return sa, formatClientID(serviceAccountID), clientSecret, nil
}

func (s *Service) LoginClientCredentials(ctx context.Context, clientID, clientSecret string) (*authentication.Token, error) {
	serviceAccountID, err := parseClientID(clientID)
	if err != nil {
		return nil, err
	}

	projectID, _, err := s.rsa.Get(ctx, serviceAccountID)
	if err != nil {
		return nil, err
	}

	ctx = auth.WithProjectID(ctx, projectID)

	us, err := s.us.GetSettings(ctx)
	if err != nil {
		return nil, err
	}

	storedPw, err := s.rsa.GetClientSecret(ctx, serviceAccountID)
	if err != nil {
		return nil, err
	}

	h := hash.New(storedPw.GetConfig())
	if err := h.Compare(clientSecret, storedPw); err != nil {
		return nil, err
	}

	sessionID := uuid.New()
	// TODO: Sessions for credentials?
	return s.generateToken(ctx, sessionID, nil, projectID, serviceAccountID, auth.SubjectTypeServiceAccount, nil, us)
}

func (s *Service) ListServiceAccounts(ctx context.Context) (iterator.Iterator[*service_account.Entry], error) {
	it, err := s.rsa.List(ctx)
	if err != nil {
		return nil, err
	}

	return iterator.Map(it, func(c *service_account.Entry) (*service_account.Entry, error) {
		sid := uuid.UUID(c.GetServiceAccountId())

		c.ClientId = formatClientID(sid)
		return c, nil
	}), nil
}

func (s *Service) DeleteServiceAccount(ctx context.Context, serviceAccountID uuid.UUID) error {
	if err := s.rsa.Delete(ctx, serviceAccountID); errors.IsNotFound(err) {
		return nil
	} else {
		return err
	}
}

func formatClientID(certificateID uuid.UUID) string {
	return fmt.Sprint("rig_", certificateID)
}

func parseClientID(clientID string) (uuid.UUID, error) {
	var certificateIDRaw string
	if n, err := fmt.Sscanf(clientID, "rig_%s", &certificateIDRaw); n != 1 || err != nil {
		return uuid.Nil, errors.InvalidArgumentErrorf("invalid client-ID format")
	}

	certificateID, err := uuid.Parse(certificateIDRaw)
	if err != nil {
		return uuid.Nil, errors.InvalidArgumentErrorf("invalid client-ID format")
	}

	return certificateID, nil
}
