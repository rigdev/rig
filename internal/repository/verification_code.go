package repository

import (
	"context"

	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig/pkg/uuid"
)

type VerificationCode interface {
	Create(ctx context.Context, verificationCode *user.VerificationCode) (*user.VerificationCode, error)
	Get(ctx context.Context, userID uuid.UUID, verificationType user.VerificationType) (*user.VerificationCode, error)
	IncreaseAttempts(ctx context.Context, userID uuid.UUID, verificationType user.VerificationType) error
	Delete(ctx context.Context, userID uuid.UUID, verificationType user.VerificationType) error
	BuildIndexes(ctx context.Context) error
}
