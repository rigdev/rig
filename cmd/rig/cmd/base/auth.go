package base

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/authentication"
	"github.com/rigdev/rig-go-api/api/v1/environment"
	"github.com/rigdev/rig-go-api/api/v1/project"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/cmdconfig"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
	"github.com/spf13/cobra"
)

var (
	OmitUser        = "OMIT_USER"
	OmitProject     = "OMIT_PROJECT"
	OmitEnvironment = "OMIT_ENVIRONMENT"
	OmitCapsule     = "OMIT_CAPSULE"
)

func CheckAuth(ctx context.Context, cmd *cobra.Command, rc rig.Client, cfg *cmdconfig.Config) error {
	if skipChecks(cmd) {
		return nil
	}

	annotations := GetAllAnnotations(cmd)

	if _, ok := annotations[OmitUser]; !ok {
		if err := authUser(ctx, rc, cfg); err != nil {
			return err
		}
	}

	if _, ok := annotations[OmitProject]; !ok {
		if err := authProject(ctx, cmd, rc, cfg); err != nil {
			return err
		}
	}
	if _, ok := annotations[OmitEnvironment]; !ok {
		if err := authEnvironment(ctx, cmd, rc, cfg); err != nil {
			return err
		}
	}

	return nil
}

func authEnvironment(ctx context.Context, cmd *cobra.Command, rig rig.Client, cfg *cmdconfig.Config) error {
	environmentID := GetEnvironment(cfg)
	if environmentID == "" {
		use, err := common.PromptConfirm("You have not selected an environment. Would you like to select one now?", true)
		if err != nil {
			return err
		}

		if !use {
			return errors.FailedPreconditionErrorf("Please select an environment or use the --environment flag")
		}

		environmentID, err = promptForEnvironment(ctx, rig)
		if err != nil {
			return err
		}
		cfg.GetCurrentContext().EnvironmentID = environmentID
		if err := cfg.Save(); err != nil {
			return err
		}
		cmd.Println("Changed environment successfully!")
	}

	res, err := rig.Environment().List(ctx, &connect.Request[environment.ListRequest]{})
	if err != nil {
		return nil
	}

	found := false
	for _, e := range res.Msg.GetEnvironments() {
		if e.GetEnvironmentId() == environmentID {
			found = true
			break
		}
	}

	if !found {
		use, err := common.PromptConfirm(
			"Your selected environment is not available. Would you like to select a new one?", true)
		if err != nil {
			return err
		}

		if !use {
			return errors.FailedPreconditionErrorf("Select an environment or use the --environment flag")
		}

		environmentID, err = promptForEnvironment(ctx, rig)
		if err != nil {
			return err
		}
		cfg.GetCurrentContext().EnvironmentID = environmentID
		if err := cfg.Save(); err != nil {
			return err
		}
		cmd.Println("Changed environment successfully!")
	}

	return nil
}

func authUser(ctx context.Context, rig rig.Client, cfg *cmdconfig.Config) error {
	user := cfg.GetCurrentAuth().UserID
	if !uuid.UUID(user).IsNil() && user != "" {
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

func authProject(ctx context.Context, cmd *cobra.Command, rig rig.Client, cfg *cmdconfig.Config) error {
	res, err := rig.Project().List(ctx, &connect.Request[project.ListRequest]{})
	if err != nil {
		return err
	}

	if len(res.Msg.Projects) == 0 {
		create, err := common.PromptConfirm("You have no projects. Would you like to create on now?", true)
		if err != nil {
			return err
		}
		if !create {
			return errors.FailedPreconditionErrorf("Create and select a project to continue")
		}

		if err := createProject(ctx, cmd, rig, cfg); err != nil {
			return err
		}

		res, err = rig.Project().List(ctx, &connect.Request[project.ListRequest]{})
		if err != nil {
			return err
		}
	}

	pid := cfg.GetCurrentContext().ProjectID
	if pid == "" || uuid.UUID(pid).IsNil() {
		use, err := common.PromptConfirm("You have not selected a project. Would you like to select one now?", true)
		if err != nil {
			return err
		}
		if !use {
			return errors.FailedPreconditionErrorf("Select a project to continue")
		}

		if err := useProject(ctx, rig, cfg); err != nil {
			return err
		}
	}

	found := false
	for _, p := range res.Msg.GetProjects() {
		if p.GetProjectId() == cfg.GetCurrentContext().ProjectID {
			found = true
			break
		}
	}

	if !found {
		// what to do here? Should we allow to use projects not existing in the
		// list? Eg. Rig project or projects form another context?
		use, err := common.PromptConfirm("Your selected project is not available. Would you like to select a new one?", true)
		if err != nil {
			return err
		}

		// if !use {
		// 	return errors.FailedPreconditionErrorf("Select a project to continue")
		// }

		if use {
			err = useProject(ctx, rig, cfg)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func createProject(ctx context.Context, cmd *cobra.Command, rc rig.Client, cfg *cmdconfig.Config) error {
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
			ProjectId:    name,
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
		cfg.GetCurrentContext().ProjectID = p.GetProjectId()
		if err := cfg.Save(); err != nil {
			return err
		}

		cmd.Println("Changed project successfully!")
	}

	return nil
}

func login(ctx context.Context, rc rig.Client, cfg *cmdconfig.Config) error {
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

func useProject(ctx context.Context, rc rig.Client, cfg *cmdconfig.Config) error {
	var projectID string
	var err error
	listRes, err := rc.Project().List(ctx, &connect.Request[project.ListRequest]{})
	if err != nil {
		return err
	}

	var ps []string
	for _, p := range listRes.Msg.GetProjects() {
		ps = append(ps, p.GetName())
	}

	i, _, err := common.PromptSelect("Project: ", ps)
	if err != nil {
		return err
	}

	projectID = listRes.Msg.GetProjects()[i].GetProjectId()

	cfg.GetCurrentContext().ProjectID = projectID
	if err := cfg.Save(); err != nil {
		return err
	}

	fmt.Println("Changed project successfully!")

	return nil
}

func promptForEnvironment(ctx context.Context, rc rig.Client) (string, error) {
	res, err := rc.Environment().List(ctx, &connect.Request[environment.ListRequest]{})
	if err != nil {
		return "", err
	}

	var es []string
	for _, e := range res.Msg.GetEnvironments() {
		es = append(es, e.GetEnvironmentId())
	}

	i, _, err := common.PromptSelect("Environment: ", es)
	if err != nil {
		return "", err
	}

	environment := res.Msg.GetEnvironments()[i].GetEnvironmentId()

	return environment, nil
}
