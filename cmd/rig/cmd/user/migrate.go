package user

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"connectrpc.com/connect"
	firebase "firebase.google.com/go"
	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/cmd/common"
	"github.com/spf13/cobra"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type migrationPlatform int32

const (
	platformUndefined migrationPlatform = iota
	platformFirebase
)

type migrationMethod int32

const (
	methodUndefined migrationMethod = iota
	methodCredentials
	methodUsersFile
)

func (m migrationPlatform) String() string {
	switch m {
	case platformFirebase:
		return "Firebase"
	default:
		return "Undefined"
	}
}

func (m migrationMethod) String() string {
	switch m {
	case methodCredentials:
		return "Credentials"
	case methodUsersFile:
		return "Users File"
	default:
		return "undefined"
	}
}

type firebaseUsers struct {
	Users []firebaseUser `json:"users"`
}

type firebaseUser struct {
	LocalID       string `json:"localId"`
	DisplayName   string `json:"displayName"`
	Email         string `json:"email"`
	PhoneNumber   string `json:"phoneNumber"`
	EmailVerified bool   `json:"emailVerified"`
	PhotoURL      string `json:"photoUrl"`
	PasswordHash  string `json:"passwordHash"`
	Salt          string `json:"salt"`
	CreatedAt     string `json:"createdAt"`
}

func (c *Cmd) migrate(ctx context.Context, _ *cobra.Command, _ []string) error {
	var err error
	fields := []string{
		platformFirebase.String(),
	}

	if platform == "" {
		_, platform, err = common.PromptSelect("Where are you migrating from?", fields)
		if err != nil {
			return err
		}
	}

	switch platform {
	case platformFirebase.String():
		if usersFilePath != "" {
			return c.migrateFromFirebaseUsersFile(ctx)
		} else if credFilePath != "" {
			return c.migrateFromFirebaseCredentials(ctx)
		} else {
			return c.migrateFromFirebase(ctx)
		}
	default:
		return fmt.Errorf("invalid migration platform")
	}
}

func (c *Cmd) migrateFromFirebase(ctx context.Context) error {
	fields := []string{
		methodCredentials.String(),
		methodUsersFile.String(),
	}

	i, _, err := common.PromptSelect("How do you want to migrate?", fields)
	if err != nil {
		return err
	}

	switch migrationMethod(i + 1) {
	case methodCredentials:
		return c.migrateFromFirebaseCredentials(ctx)
	case methodUsersFile:
		return c.migrateFromFirebaseUsersFile(ctx)
	default:
		return fmt.Errorf("invalid migration method")
	}
}

func (c *Cmd) migrateFromFirebaseCredentials(ctx context.Context) error {
	var err error
	if credFilePath == "" {
		credFilePath, err = common.PromptInput("Credentials Path:", common.ValidateNonEmptyOpt)
		if err != nil {
			return err
		}
	}

	file, err := os.Open(credFilePath)
	if err != nil {
		return err
	}

	bytevalue, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	cc := map[string]interface{}{}
	if err := json.Unmarshal(bytevalue, &cc); err != nil {
		return err
	}

	projectID, ok := cc["project_id"].(string)
	if !ok {
		return fmt.Errorf("project_id not found in credentials")
	}

	app, err := firebase.NewApp(ctx, &firebase.Config{ProjectID: projectID}, option.WithCredentialsJSON(bytevalue))
	if err != nil {
		return err
	}

	authClient, err := app.Auth(ctx)
	if err != nil {
		return err
	}

	// input hashing key for password
	if hashingKey == "" {
		hashingKey, err = common.PromptInput("Hashing Key:", common.ValidateNonEmptyOpt)
		if err != nil {
			return err
		}
	}

	hashingConfig := &model.HashingConfig{
		Method: &model.HashingConfig_Scrypt{
			Scrypt: &model.ScryptHashingConfig{
				SignerKey:     string(hashingKey),
				SaltSeparator: string("Bw=="),
				Rounds:        8,
				MemCost:       14,
				P:             1,
				KeyLen:        int32(32),
			},
		},
	}

	numUsersMigrated := 0
	errors := make(map[string]error)

	iter := authClient.Users(ctx, "")
	for {
		u, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		if u.PasswordHash == "" {
			return fmt.Errorf("user has no password, or credentials dont have access")
		}

		// Decode salt and pw hash in url encoding
		salt, err := base64.URLEncoding.DecodeString(u.PasswordSalt)
		if err != nil {
			return err
		}
		fmt.Println(u.PasswordHash)
		passwordHash, err := base64.URLEncoding.DecodeString(u.PasswordHash)
		if err != nil {
			return err
		}

		us := []*user.Update{
			{
				Field: &user.Update_HashedPassword{
					HashedPassword: &model.HashingInstance{
						Config: hashingConfig,
						Hash:   passwordHash,
						Instance: &model.HashingInstance_Scrypt{
							Scrypt: &model.ScryptHashingInstance{
								Salt: salt,
							},
						},
					},
				},
			},
		}

		if u.Email != "" {
			us = append(us, &user.Update{
				Field: &user.Update_Email{
					Email: u.Email,
				},
			})
		}

		if u.PhoneNumber != "" {
			us = append(us, &user.Update{
				Field: &user.Update_PhoneNumber{
					PhoneNumber: u.PhoneNumber,
				},
			})
		}

		_, err = c.Rig.User().Create(ctx, &connect.Request[user.CreateRequest]{
			Msg: &user.CreateRequest{
				Initializers: us,
			},
		})
		if err != nil {
			errors[u.Email] = err
		} else {
			numUsersMigrated++
		}
	}

	fmt.Printf("Successfully migrated %v users \n", numUsersMigrated)
	if len(errors) > 0 {
		fmt.Println("Errors:")
		for email, err := range errors {
			fmt.Printf("%v: %v\n", email, err)
		}
	}
	return nil
}

func (c *Cmd) migrateFromFirebaseUsersFile(ctx context.Context) error {
	var err error
	if usersFilePath == "" {
		usersFilePath, err = common.PromptInput("users.json path:", common.ValidateNonEmptyOpt)
		if err != nil {
			return err
		}
	}

	// load json credentials file from path
	jsonFile, err := os.Open(usersFilePath)
	if err != nil {
		return err
	}
	defer jsonFile.Close()

	bytevalue, err := io.ReadAll(jsonFile)
	if err != nil {
		return err
	}

	var users firebaseUsers
	err = json.Unmarshal(bytevalue, &users)
	if err != nil {
		return err
	}
	fmt.Println("Successfully parsed", len(users.Users), "users")

	// input hashing key for password
	if hashingKey == "" {
		hashingKey, err = common.PromptInput("Hashing Key:", common.ValidateNonEmptyOpt)
		if err != nil {
			return err
		}
	}

	hashingConfig := &model.HashingConfig{
		Method: &model.HashingConfig_Scrypt{
			Scrypt: &model.ScryptHashingConfig{
				SignerKey:     hashingKey,
				SaltSeparator: "Bw==",
				Rounds:        8,
				MemCost:       14,
				P:             1,
				KeyLen:        int32(32),
			},
		},
	}

	numUsersMigrated := 0
	errors := make(map[string]error)
	for _, fu := range users.Users {
		// base64 decode salt and password hash
		salt, err := base64.StdEncoding.DecodeString(fu.Salt)
		if err != nil {
			return err
		}
		passwordHash, err := base64.StdEncoding.DecodeString(fu.PasswordHash)
		if err != nil {
			return err
		}

		us := []*user.Update{
			{
				Field: &user.Update_HashedPassword{
					HashedPassword: &model.HashingInstance{
						Config: hashingConfig,
						Hash:   passwordHash,
						Instance: &model.HashingInstance_Scrypt{
							Scrypt: &model.ScryptHashingInstance{
								Salt: salt,
							},
						},
					},
				},
			},
		}

		if fu.Email != "" {
			us = append(us, &user.Update{
				Field: &user.Update_Email{
					Email: fu.Email,
				},
			})
		}

		if fu.PhoneNumber != "" {
			us = append(us, &user.Update{
				Field: &user.Update_PhoneNumber{
					PhoneNumber: fu.PhoneNumber,
				},
			})
		}

		_, err = c.Rig.User().Create(ctx, &connect.Request[user.CreateRequest]{
			Msg: &user.CreateRequest{
				Initializers: us,
			},
		})
		if err != nil {
			errors[fu.Email] = err
		} else {
			numUsersMigrated++
		}
	}

	fmt.Printf("Successfully migrated %v users \n", numUsersMigrated)
	if len(errors) > 0 {
		fmt.Println("Errors:")
		for email, err := range errors {
			fmt.Printf("%v: %v\n", email, err)
		}
	}
	return nil
}
