package authentication

import (
	"context"
	"reflect"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/authentication"
	"github.com/rigdev/rig-go-api/api/v1/user"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
)

// Register inserts new user in the namespace.
func (h *Handler) Register(ctx context.Context, req *connect.Request[authentication.RegisterRequest]) (*connect.Response[authentication.RegisterResponse], error) {
	var inits []*user.Update

	switch v := req.Msg.GetMethod().(type) {
	case *authentication.RegisterRequest_UserPassword:
		pID, err := uuid.Parse(v.UserPassword.GetProjectId())
		if err != nil {
			return nil, errors.InvalidArgumentErrorf("invalid project ID")
		}
		ctx = auth.WithProjectID(ctx, pID)

		inits = append(inits, &user.Update{
			Field: &user.Update_Password{
				Password: v.UserPassword.Password,
			},
		})

		var lt model.LoginType
		switch i := v.UserPassword.GetIdentifier().GetIdentifier().(type) {
		case *model.UserIdentifier_Email:
			lt = model.LoginType_LOGIN_TYPE_EMAIL_PASSWORD
			inits = append(inits, &user.Update{
				Field: &user.Update_Email{Email: i.Email},
			})
		case *model.UserIdentifier_Username:
			lt = model.LoginType_LOGIN_TYPE_USERNAME_PASSWORD
			inits = append(inits, &user.Update{
				Field: &user.Update_Username{Username: i.Username},
			})
		case *model.UserIdentifier_PhoneNumber:
			lt = model.LoginType_LOGIN_TYPE_PHONE_PASSWORD
			inits = append(inits, &user.Update{
				Field: &user.Update_PhoneNumber{PhoneNumber: i.PhoneNumber},
			})
		default:
			return nil, errors.InvalidArgumentErrorf("invalid identifier type '%v'", reflect.TypeOf(i))
		}

		if _, err := h.us.CreateUser(ctx, &model.RegisterMethod{
			Method: &model.RegisterMethod_Signup_{
				Signup: &model.RegisterMethod_Signup{LoginType: lt},
			},
		}, inits); err != nil {
			return nil, err
		}

		userID, u, t, err := h.as.LoginUserPassword(ctx, req.Msg.GetUserPassword().GetIdentifier(), v.UserPassword.GetPassword())
		if err != nil {
			return nil, err
		}
		return &connect.Response[authentication.RegisterResponse]{
			Msg: &authentication.RegisterResponse{
				Token:    t,
				UserId:   userID.String(),
				UserInfo: u,
			},
		}, nil

	default:
		return nil, errors.InvalidArgumentErrorf("invalid register method '%v'", reflect.TypeOf(v))
	}
}
