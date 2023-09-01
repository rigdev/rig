package crypto

import (
	"crypto/rand"
	"math/big"
)

var (
	AlphaNum      = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	Alpha         = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	AlphaLowerNum = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	AlphaUpperNum = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	AlphaLower    = []rune("abcdefghijklmnopqrstuvwxyz")
	AlphaUpper    = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	Numeric       = []rune("0123456789")
)

func GenerateSymmetricKey(length int, runes []rune) (string, error) {
	b := make([]rune, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(runes))))
		if err != nil {
			return "", err
		}
		b[i] = runes[n.Uint64()]
	}
	return string(b), nil
}
