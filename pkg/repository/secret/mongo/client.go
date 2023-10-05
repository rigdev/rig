package mongo

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha512"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/pbkdf2"
)

type MongoRepository struct {
	Collection *mongo.Collection
	key        []byte
	gcm        cipher.AEAD
}

func NewRepository(c *mongo.Client, key string) (*MongoRepository, error) {
	aesKey := pbkdf2.Key(
		[]byte(key),
		[]byte{0xe5, 0x5e, 0x59, 0x0f, 0xe9, 0x3e, 0x40, 0x17, 0xb5, 0x3b, 0xb8, 0x5a, 0x59, 0x1b, 0x8c, 0x21},
		210000,
		32,
		sha512.New,
	)

	aes, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(aes)
	if err != nil {
		return nil, err
	}

	repo := &MongoRepository{
		Collection: c.Database("rig").Collection("secrets"),
		gcm:        gcm,
	}
	if err := repo.BuildIndexes(context.Background()); err != nil {
		return nil, err
	}
	return repo, nil
}

// BuildIndexes builds the indexes for the Mongo Object.
func (r *MongoRepository) BuildIndexes(ctx context.Context) error {
	projectIdIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "project_id", Value: 1},
			{Key: "secret_id", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}
	if _, err := r.Collection.Indexes().CreateOne(ctx, projectIdIndexModel); err != nil {
		return err
	}
	return nil
}

func (r *MongoRepository) encryptSecret(secret []byte) ([]byte, error) {
	nonce := make([]byte, r.gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	return r.gcm.Seal(nonce, nonce, secret, nil), nil
}

func (r *MongoRepository) decryptSecret(c []byte) ([]byte, error) {
	nonceSize := r.gcm.NonceSize()
	if len(c) < nonceSize {
		return nil, fmt.Errorf("invalid ciphertext")
	}
	nonce, ciphertext := c[:nonceSize], c[nonceSize:]

	return r.gcm.Open(nil, nonce, ciphertext, nil)
}
