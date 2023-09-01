// Package hash is build to support multiple hashing algorithms with different parameters across projects.
package hash

import (
	"fmt"
	"reflect"

	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/pkg/errors"
)

type Hash interface {
	Generate(password string) (*model.HashingInstance, error)
	Compare(password string, hashModel *model.HashingInstance) error
}

type hasher struct {
	cfg *model.HashingConfig
}

func New(config *model.HashingConfig) Hash {
	return &hasher{cfg: config}
}

func (h *hasher) Generate(password string) (*model.HashingInstance, error) {
	if h.cfg == nil {
		return nil, errors.InvalidArgumentErrorf("missing password hashing config")
	}

	switch v := h.cfg.GetMethod().(type) {
	case *model.HashingConfig_Bcrypt:
		if v.Bcrypt.GetCost() == 0 {
			return nil, errors.InvalidArgumentErrorf("invalid bcrypt arguments")
		}
		return (&bcryptHash{cfg: v.Bcrypt}).Generate(password)
	case *model.HashingConfig_Scrypt:
		return (&scryptHash{
			cfg: v.Scrypt,
		}).generate(password)
	default:
		return nil, errors.InvalidArgumentErrorf("invalid password hashing method")
	}
}

func (h *hasher) Compare(password string, hash *model.HashingInstance) error {
	if password == "" {
		return errors.InvalidArgumentErrorf("password is empty")
	}
	if hash == nil {
		return errors.InvalidArgumentErrorf("invalid password comparison target")
	}
	switch v := hash.GetInstance().(type) {
	case *model.HashingInstance_Bcrypt:
		return (&bcryptHash{cfg: h.cfg.GetBcrypt()}).Compare(password, string(hash.GetHash()))
	case *model.HashingInstance_Scrypt:
		return (&scryptHash{cfg: h.cfg.GetScrypt()}).compare(password, v.Scrypt.GetSalt(), hash.GetHash())
	default:
		return fmt.Errorf("hash compare: invalid hash %s", reflect.TypeOf(v))
	}
}
