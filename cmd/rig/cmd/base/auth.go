package base

import (
	"context"
	"fmt"
	"strings"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/authentication"
	"github.com/rigdev/rig-go-api/api/v1/project"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/cmd_config"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	OmitUser    = "OMIT_USER"
	OmitProject = "OMIT_PROJECT"
)

func CheckAuth(cmd *cobra.Command, rc rig.Client, cfg *cmd_config.Config, logger *zap.Logger) error {
	if _, ok := cmd.Annotations[OmitUser]; ok {
		return nil
	} else if uuid.UUID(cfg.GetCurrentAuth().UserID).IsNil() {
		login, err := common.PromptConfirm("You are not logged in. Would you like to login now?", true)
		if err != nil {
			return err
		}
		if !login {
			return errors.UnauthenticatedErrorf("Login to continue")
		}
		err = Login(rc, cfg)
		if err != nil {
			return err
		}
	}

	if _, ok := cmd.Annotations[OmitProject]; ok {
		return nil
	} else if uuid.UUID(cfg.GetCurrentContext().Project.ProjectID).IsNil() {
		use, err := common.PromptConfirm("You have not selected a project. Would you like to select one now?", true)
		if err != nil {
			return err
		}
		if !use {
			return errors.FailedPreconditionErrorf("Select a project to continue")
		}
		err = UseProject(rc, cfg)
		if err != nil {
			return err
		}
	}
	return nil
}

func Login(rc rig.Client, cfg *cmd_config.Config) error {
	u, err := common.PromptGetInput("Enter Username or Email", common.ValidateNonEmpty)
	if err != nil {
		return err
	}

	var id *model.UserIdentifier
	if strings.Contains(u, "@") {
		id = &model.UserIdentifier{
			Identifier: &model.UserIdentifier_Email{
				Email: u,
			},
		}
	} else {
		id = &model.UserIdentifier{
			Identifier: &model.UserIdentifier_Username{
				Username: u,
			},
		}
	}

	pw, err := common.GetPasswordPrompt("Enter Password")
	if err != nil {
		return err
	}

	res, err := rc.Authentication().Login(context.Background(), &connect.Request[authentication.LoginRequest]{
		Msg: &authentication.LoginRequest{
			Method: &authentication.LoginRequest_UserPassword{
				UserPassword: &authentication.UserPassword{
					Identifier: id,
					Password:   pw,
					ProjectId:  auth.RigProjectID.String(),
				},
			},
		},
	})
	if err != nil {
		return err
	}

	uid, err := uuid.Parse(res.Msg.GetUserId())
	if err != nil {
		return err
	}

	cfg.GetCurrentAuth().UserID = uid
	cfg.GetCurrentAuth().AccessToken = res.Msg.GetToken().GetAccessToken()
	cfg.GetCurrentAuth().RefreshToken = res.Msg.GetToken().GetRefreshToken()
	if err := cfg.Save(); err != nil {
		return err
	}

	fmt.Println("Login successful!")
	return nil
}

func UseProject(rc rig.Client, cfg *cmd_config.Config) error {
	var projectID uuid.UUID
	var err error
	ctx := context.Background()
	list_res, err := rc.Project().List(ctx, &connect.Request[project.ListRequest]{})
	if err != nil {
		return err
	}

	var ps []string
	for _, p := range list_res.Msg.GetProjects() {
		ps = append(ps, p.GetName())
	}

	i, _, err := common.PromptSelect("Project: ", ps, false)
	if err != nil {
		return err
	}

	projectID, err = uuid.Parse(list_res.Msg.GetProjects()[i].GetProjectId())
	if err != nil {
		return err
	}

	res, err := rc.Project().Use(ctx, &connect.Request[project.UseRequest]{
		Msg: &project.UseRequest{
			ProjectId: projectID.String(),
		},
	})
	if err != nil {
		return err
	}

	cfg.GetCurrentContext().Project.ProjectID = projectID
	cfg.GetCurrentContext().Project.ProjectToken = res.Msg.GetProjectToken()
	if err := cfg.Save(); err != nil {
		return err
	}

	fmt.Println("Changed project successfully!")

	return nil
}
