package hash

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/pkg/errors"
	"golang.org/x/crypto/scrypt"
)

type scryptHash struct {
	cfg *model.ScryptHashingConfig
}

var DefaultScrypt = &model.ScryptHashingConfig{
	SignerKey:     "",
	SaltSeparator: "Bw==",
	Rounds:        8,
	MemCost:       14,
	P:             1,
	KeyLen:        32,
}

func (sh *scryptHash) generate(password string) (*model.HashingInstance, error) {
	salt := make([]byte, 12)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}
	hash, err := Key([]byte(password), salt, sh.cfg)
	if err != nil {
		return nil, err
	}
	return &model.HashingInstance{
		Config: &model.HashingConfig{
			Method: &model.HashingConfig_Scrypt{
				Scrypt: sh.cfg,
			},
		},
		Hash: hash,
		Instance: &model.HashingInstance_Scrypt{
			Scrypt: &model.ScryptHashingInstance{
				Salt: salt,
			},
		},
	}, nil
}

func (sh *scryptHash) compare(password string, salt, hash []byte) error {
	hs, err := sh.encode([]byte(password), salt)
	if err != nil {
		return err
	}
	if bytes.Equal(hs, hash) {
		return nil
	}
	return errors.UnauthenticatedErrorf("wrong password")
}

func (sh *scryptHash) encodeToString(password, salt []byte) (string, error) {
	res, err := sh.encode(password, salt)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(res), nil
}

func (sh *scryptHash) encode(password, salt []byte) ([]byte, error) {
	signerKey, err := base64.StdEncoding.DecodeString(string(sh.cfg.GetSignerKey()))
	if err != nil {
		fmt.Println("ERROR IN KEY")
		return nil, err
	}
	saltSeparator, err := base64.StdEncoding.DecodeString(string(sh.cfg.GetSaltSeparator()))
	if err != nil {
		fmt.Println("ERROR IN SALT")
		return nil, err
	}
	return key(password, salt, signerKey, saltSeparator, int(sh.cfg.GetRounds()), int(sh.cfg.GetMemCost()), int(sh.cfg.GetP()), int(sh.cfg.GetKeyLen()))
}

func Key(password, salt []byte, cfg *model.ScryptHashingConfig) ([]byte, error) {
	var (
		sk, ss []byte
		err    error
	)
	if sk, err = base64.StdEncoding.DecodeString(cfg.GetSignerKey()); err != nil {
		return nil, err
	}
	if ss, err = base64.StdEncoding.DecodeString(cfg.GetSaltSeparator()); err != nil {
		return nil, err
	}
	return key(password, salt, sk, ss, int(cfg.GetRounds()), int(cfg.GetMemCost()), int(cfg.GetP()), int(cfg.GetKeyLen()))
}

func key(password, salt, signerKey, saltSeparator []byte, rounds, memCost, p, keyLen int) ([]byte, error) {
	ck, err := scrypt.Key(password, append(salt, saltSeparator...), 1<<memCost, rounds, p, keyLen)
	if err != nil {
		fmt.Println("ERROR IN KEYGEN")
		return nil, err
	}
	var block cipher.Block
	if block, err = aes.NewCipher(ck); err != nil {
		fmt.Println("ERROR IN KEYGEN")
		return nil, err
	}
	cipherText := make([]byte, aes.BlockSize+len(signerKey))
	stream := cipher.NewCTR(block, cipherText[:aes.BlockSize])
	stream.XORKeyStream(cipherText[aes.BlockSize:], signerKey)
	return cipherText[aes.BlockSize:], nil
}
