package hash

import (
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type bcryptHash struct {
	cfg *model.BcryptHashingConfig
}

const (
	BcryptMinCost     = bcrypt.MinCost
	BcryptMaxCost     = bcrypt.MaxCost
	BcryptDefaultCost = bcrypt.DefaultCost
)

var (
	ErrBcryptCostMinValue = errors.InvalidArgumentErrorf("bcrypt cost myst be a minimum of %d", BcryptMinCost)
	ErrBcryptCostMaxValue = errors.InvalidArgumentErrorf("bcrypt cost myst be a maximum of %d", BcryptMaxCost)
)

var DefaultBcrypt = &model.BcryptHashingConfig{
	Cost: int32(BcryptDefaultCost),
}

func (bh *bcryptHash) Generate(password string) (*model.HashingInstance, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), int(bh.cfg.GetCost()))
	if err != nil {
		return nil, err
	}
	return &model.HashingInstance{
		Config: &model.HashingConfig{
			Method: &model.HashingConfig_Bcrypt{
				Bcrypt: bh.cfg,
			},
		},
		Hash:     hash,
		Instance: &model.HashingInstance_Bcrypt{Bcrypt: &model.BcryptHashingInstance{}},
	}, nil
}

func (bh *bcryptHash) Compare(password, hash string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err == bcrypt.ErrMismatchedHashAndPassword {
		return errors.UnauthenticatedErrorf("wrong password")
	} else {
		return err
	}
}
