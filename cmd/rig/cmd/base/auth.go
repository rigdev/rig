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
)

var (
	OmitUser    = "OMIT_USER"
	OmitProject = "OMIT_PROJECT"
)

func CheckAuth(cmd *cobra.Command, rc rig.Client, cfg *cmd_config.Config) error {
	ctx := context.Background()

	if _, ok := cmd.Annotations[OmitUser]; !ok {
		if err := authUser(ctx, rc, cfg); err != nil {
			return err
		}
	}

	if _, ok := cmd.Annotations[OmitProject]; !ok {
		if err := authProject(ctx, cmd, rc, cfg); err != nil {
			return err
		}
	}

	return nil
}

func authUser(ctx context.Context, rig rig.Client, cfg *cmd_config.Config) error {
	if cfg.GetCurrentAuth().UserID != "" {
		return nil
	}
	loginBool, err := common.PromptConfirm("You are not logged in. Would you like to login now?", true)
	if err != nil {
		return err
	}
	if !loginBool {
		return errors.UnauthenticatedErrorf("Login to continue")
	}
	return login(ctx, rig, cfg)
}

func authProject(ctx context.Context, cmd *cobra.Command, rig rig.Client, cfg *cmd_config.Config) error {
	res, err := rig.Project().List(ctx, &connect.Request[project.ListRequest]{})
	if err != nil {
		return err
	}

	if len(res.Msg.GetProjects()) == 0 {
		create, err := common.PromptConfirm("You have no projects. Would you like to create on now?", true)
		if err != nil {
			return err
		}
		if !create {
			return errors.FailedPreconditionErrorf("Create and select a project to continue")
		}

		err = createProject(ctx, cmd, rig, cfg)
		if err != nil {
			return err
		}
	}

	if cfg.GetCurrentContext().Project.ProjectID == "" {
		use, err := common.PromptConfirm("You have not selected a project. Would you like to select one now?", true)
		if err != nil {
			return err
		}
		if !use {
			return errors.FailedPreconditionErrorf("Select a project to continue")
		}

		err = useProject(ctx, rig, cfg)
		if err != nil {
			return err
		}
	}

	found := false
	for _, p := range res.Msg.GetProjects() {
		if p.GetProjectId() == cfg.GetCurrentContext().Project.ProjectID {
			found = true
			break
		}
	}

	if !found {
		use, err := common.PromptConfirm("Your selected project is not available. Would you like to select a new one?", true)
		if err != nil {
			return err
		}
		if !use {
			return errors.FailedPreconditionErrorf("Select a project to continue")
		}

		err = useProject(ctx, rig, cfg)
		if err != nil {
			return err
		}
	}
	return nil
}

func createProject(ctx context.Context, cmd *cobra.Command, rc rig.Client, cfg *cmd_config.Config) error {
	name, err := common.PromptInput("Project name:", common.ValidateNonEmptyOpt)
	if err != nil {
		return err
	}

	initializers := []*project.Update{
		{
			Field: &project.Update_Name{
				Name: name,
			},
		},
	}

	res, err := rc.Project().Create(ctx, &connect.Request[project.CreateRequest]{
		Msg: &project.CreateRequest{
			Initializers: initializers,
		},
	})
	if err != nil {
		return err
	}

	p := res.Msg.GetProject()
	cmd.Printf("Successfully created project %s with id %s \n", name, p.GetProjectId())

	useProject, err := common.PromptConfirm("Would you like to use this project now?", true)
	if err != nil {
		return err
	}

	if useProject {
		res, err := rc.Project().Use(ctx, &connect.Request[project.UseRequest]{
			Msg: &project.UseRequest{
				ProjectId: p.GetProjectId(),
			},
		})
		if err != nil {
			return err
		}

		cfg.GetCurrentContext().Project.ProjectID = p.GetProjectId()
		cfg.GetCurrentContext().Project.ProjectToken = res.Msg.GetProjectToken()
		if err := cfg.Save(); err != nil {
			return err
		}

		cmd.Println("Changed project successfully!")
	}

	return nil
}

func login(ctx context.Context, rc rig.Client, cfg *cmd_config.Config) error {
	u, err := common.PromptInput("Enter Username or Email:", common.ValidateNonEmptyOpt)
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

	pw, err := common.PromptPassword("Enter Password")
	if err != nil {
		return err
	}

	res, err := rc.Authentication().Login(ctx, &connect.Request[authentication.LoginRequest]{
		Msg: &authentication.LoginRequest{
			Method: &authentication.LoginRequest_UserPassword{
				UserPassword: &authentication.UserPassword{
					Identifier: id,
					Password:   pw,
					ProjectId:  auth.RigProjectID,
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

func useProject(ctx context.Context, rc rig.Client, cfg *cmd_config.Config) error {
	var projectID string
	var err error
	list_res, err := rc.Project().List(ctx, &connect.Request[project.ListRequest]{})
	if err != nil {
		return err
	}

	var ps []string
	for _, p := range list_res.Msg.GetProjects() {
		ps = append(ps, p.GetName())
	}

	i, _, err := common.PromptSelect("Project: ", ps)
	if err != nil {
		return err
	}

	projectID = list_res.Msg.GetProjects()[i].GetProjectId()

	res, err := rc.Project().Use(ctx, &connect.Request[project.UseRequest]{
		Msg: &project.UseRequest{
			ProjectId: projectID,
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
