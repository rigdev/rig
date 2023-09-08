package cmd

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/lucasepe/codename"
	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/internal/config"
	auth_service "github.com/rigdev/rig/internal/service/auth"
	user_service "github.com/rigdev/rig/internal/service/user"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/uuid"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	userEmail       string
	userUsername    string
	userPhoneNumber string
	userPassword    string
	userCount       int
)

func init() {
	users := &cobra.Command{
		Use: "users",
	}

	create := &cobra.Command{
		Use:  "create",
		RunE: register(UsersCreate),
	}
	create.PersistentFlags().StringVar(&userEmail, "email", "", "email for the user")
	create.PersistentFlags().StringVar(&userUsername, "username", "", "username for the user")
	create.PersistentFlags().StringVar(&userPhoneNumber, "phone", "", "phone number for the user")
	create.PersistentFlags().StringVar(&userPassword, "password", "", "password for the user")
	users.AddCommand(create)

	lookup := &cobra.Command{
		Use:  "lookup",
		RunE: register(UsersLookup),
	}
	lookup.PersistentFlags().StringVar(&userEmail, "email", "", "email for the user")
	lookup.PersistentFlags().StringVar(&userUsername, "username", "", "username for the user")
	users.AddCommand(lookup)

	populate := &cobra.Command{
		Use:  "populate",
		RunE: register(UsersPopulate),
	}
	populate.PersistentFlags().IntVarP(&userCount, "count", "n", 1, "number of users to create")
	populate.PersistentFlags().StringVar(&userUsername, "username", "", "username for the user")
	users.AddCommand(populate)

	verifyEmail := &cobra.Command{
		Use:  "verify-email email",
		Args: cobra.ExactArgs(1),
		RunE: register(UsersVerifyEmail),
	}
	users.AddCommand(verifyEmail)

	resetSessions := &cobra.Command{
		Use:  "reset-sessions user-id",
		Args: cobra.ExactArgs(1),
		RunE: register(UsersResetSessions),
	}
	users.AddCommand(resetSessions)

	list := &cobra.Command{
		Use:  "list [search]",
		RunE: register(UsersList),
	}
	users.AddCommand(list)

	delete := &cobra.Command{
		Use:  "delete",
		RunE: register(UsersDelete),
	}
	users.AddCommand(delete)

	listSessions := &cobra.Command{
		Use:  "list-sessions <user-id>",
		Args: cobra.ExactArgs(1),
		RunE: register(UsersListSessions),
	}
	users.AddCommand(listSessions)

	getSettings := &cobra.Command{
		Use:  "get-settings",
		RunE: register(UsersGetSettings),
	}
	users.AddCommand(getSettings)

	rootCmd.AddCommand(users)
}

func UsersLookup(ctx context.Context, cmd *cobra.Command, us user_service.Service, cfg config.Config, logger *zap.Logger) error {
	uID := &model.UserIdentifier{}

	if userEmail != "" {
		uID.Identifier = &model.UserIdentifier_Email{Email: userEmail}
	} else if userUsername != "" {
		uID.Identifier = &model.UserIdentifier_Username{Username: userUsername}
	}

	u, err := us.GetUserByIdentifier(ctx, uID)
	if err != nil {
		return err
	}

	logger.Info("found user", zap.String("user_id", u.GetUserId()), zap.String("email", u.GetUserInfo().GetEmail()), zap.String("username", u.GetUserInfo().GetUsername()))

	return nil
}

func UsersCreate(ctx context.Context, cmd *cobra.Command, us user_service.Service, logger *zap.Logger) error {
	// This helps bypass register-disallowed.
	ctx = auth.WithClaims(ctx, auth_service.ProjectClaims{
		UseProjectID: auth.RigProjectID,
	})

	if userPassword == "" {
		pw, err := common.PromptPassword("Password:")
		if err != nil {
			return err
		}

		userPassword = pw
	}

	var ups []*user.Update

	if userEmail != "" {
		ups = append(ups, &user.Update{Field: &user.Update_Email{Email: userEmail}})
	}

	if userUsername != "" {
		ups = append(ups, &user.Update{Field: &user.Update_Username{Username: userUsername}})
	}

	if userPhoneNumber != "" {
		ups = append(ups, &user.Update{Field: &user.Update_PhoneNumber{PhoneNumber: userPhoneNumber}})
	}

	if userPassword != "" {
		ups = append(ups, &user.Update{Field: &user.Update_Password{Password: userPassword}})
	}

	u, err := us.CreateUser(ctx, &model.RegisterMethod{Method: &model.RegisterMethod_System_{}}, ups)
	if err != nil {
		return err
	}

	logger.Info("created user", zap.String("user_id", u.GetUserId()), zap.String("email", u.GetUserInfo().GetEmail()), zap.String("username", u.GetUserInfo().GetUsername()))

	return nil
}

func UsersPopulate(ctx context.Context, cmd *cobra.Command, us user_service.Service, logger *zap.Logger) error {
	logger.Info("created users", zap.Int("count", userCount))

	ctx = auth.WithClaims(ctx, auth_service.ProjectClaims{
		UseProjectID: auth.RigProjectID,
	})
	for i := 0; i < userCount; i++ {

		var ups []*user.Update

		rng, err := codename.DefaultRNG()
		if err != nil {
			return err
		}

		ups = append(ups, &user.Update{Field: &user.Update_Username{Username: strings.ToLower(codename.Generate(rng, 8))}})

		ups = append(ups, &user.Update{Field: &user.Update_Password{Password: fmt.Sprint(codename.Generate(rng, 0), "4!")}})

		ups = append(ups, &user.Update{Field: &user.Update_Email{
			Email: fmt.Sprint(strings.ToLower(codename.Generate(rng, 8)), "@dummy.email"),
		}})

		ups = append(ups, &user.Update{Field: &user.Update_Profile{
			Profile: &user.Profile{
				FirstName: codename.Generate(rng, 0),
				LastName:  codename.Generate(rng, 0),
			},
		}})

		u, err := us.CreateUser(ctx, &model.RegisterMethod{Method: &model.RegisterMethod_System_{}}, ups)
		if err != nil {
			return err
		}

		logger.Info("created user", zap.String("user_id", u.GetUserId()), zap.String("email", u.GetUserInfo().GetEmail()), zap.String("username", u.GetUserInfo().GetUsername()))

	}
	return nil
}

func UsersVerifyEmail(ctx context.Context, cmd *cobra.Command, args []string, us user_service.Service, logger *zap.Logger) error {
	uID := &model.UserIdentifier{
		Identifier: &model.UserIdentifier_Email{Email: args[0]},
	}

	u, err := us.GetUserByIdentifier(ctx, uID)
	if err != nil {
		return err
	}

	if err := us.UpdateUser(ctx, uuid.UUID(u.GetUserId()), []*user.Update{{
		Field: &user.Update_IsEmailVerified{
			IsEmailVerified: true,
		},
	}}); err != nil {
		return err
	}

	logger.Info("email verified", zap.String("user_id", u.GetUserId()), zap.String("email", u.GetUserInfo().GetEmail()))

	return nil
}

func UsersList(ctx context.Context, cmd *cobra.Command, args []string, us user_service.Service, logger *zap.Logger) error {
	search := strings.Join(args, " ")
	it, _, err := us.List(ctx, &model.Pagination{}, search)
	if err != nil {
		return err
	}

	defer it.Close()

	for {
		u, err := it.Next()
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}

		logger.Info("found user", zap.String("name", u.GetPrintableName()), zap.String("user_id", u.GetUserId()))
	}
}

func UsersDelete(ctx context.Context, cmd *cobra.Command, args []string, us user_service.Service, logger *zap.Logger) error {
	userID, err := uuid.Parse(args[0])
	if err != nil {
		return err
	}

	if err := us.DeleteUser(ctx, userID); err != nil {
		return err
	}

	logger.Info("deleted user", zap.Stringer("user_id", userID))

	return nil
}

func UsersResetSessions(ctx context.Context, cmd *cobra.Command, args []string, us user_service.Service, logger *zap.Logger) error {
	userID, err := uuid.Parse(args[0])
	if err != nil {
		return err
	}

	if err := us.UpdateUser(ctx, userID, []*user.Update{{
		Field: &user.Update_ResetSessions_{
			ResetSessions: &user.Update_ResetSessions{},
		},
	}}); err != nil {
		return err
	}

	logger.Info("user sessions reset", zap.Stringer("user_id", userID))

	return nil
}

func UsersListSessions(ctx context.Context, cmd *cobra.Command, args []string, as *auth_service.Service, logger *zap.Logger) error {
	userID, err := uuid.Parse(args[0])
	if err != nil {
		return err
	}

	it, err := as.ListSessions(ctx, userID)
	if err != nil {
		return err
	}

	defer it.Close()

	for {
		s, err := it.Next()
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}

		logger.Info("session", zap.String("session_id", s.GetSessionId()), zap.Time("expires_at", s.GetSession().GetExpiresAt().AsTime()))
	}
}

func UsersGetSettings(ctx context.Context, cmd *cobra.Command, us user_service.Service, logger *zap.Logger) error {
	settings, err := us.GetSettings(ctx)
	if err != nil {
		return err
	}
	logger.Info("settings", zap.Any("settings", settings))
	return nil
}
