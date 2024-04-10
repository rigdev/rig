package auth

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
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/rigdev/rig/pkg/cli/scope"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
)

type Service struct {
	rig      rig.Client
	scope    scope.Scope
	prompter common.Prompter
}

func NewService(rig rig.Client, scope scope.Scope, prompter common.Prompter) *Service {
	return &Service{
		rig:      rig,
		scope:    scope,
		prompter: prompter,
	}
}

var (
	OmitUser        = "OMIT_USER"
	OmitProject     = "OMIT_PROJECT"
	OmitEnvironment = "OMIT_ENVIRONMENT"
	OmitCapsule     = "OMIT_CAPSULE"
)

func (s *Service) CheckAuth(ctx context.Context, interactive bool, f func(context.Context, bool) error) error {
	for {
		retry, err := s.handleAuthError(f(ctx, interactive), interactive)
		if err != nil {
			return err
		}
		if retry {
			continue
		}
		return nil
	}
}

func (s *Service) handleAuthError(origErr error, interactive bool) (bool, error) {
	if !errors.IsUnauthenticated(origErr) {
		return false, origErr
	}

	if strings.Contains(origErr.Error(), "wrong password") {
		fmt.Println("Wrong username or password.")
		return true, nil
	}

	cmdContext := s.scope.GetCurrentContext()
	str := fmt.Sprintf(
		"There seems to be an issue with the authentication information stored in your current context '%s'",
		cmdContext.Name,
	)
	if !interactive {
		return false, fmt.Errorf("%s: %s", str, origErr)
	}
	fmt.Println(str)
	ok, err := s.prompter.Confirm("Do you wish to rebuild this context before proceeding?", true)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, origErr
	}

	s.scope.GetCfg().DeleteContext(cmdContext.Name)
	if err := s.scope.GetCfg().CreateContext(cmdContext.Name, cmdContext.GetService().Server, interactive); err != nil {
		return false, err
	}

	return true, nil
}

func (s *Service) AuthEnvironment(ctx context.Context, interactive bool) error {
	environmentID := flags.GetEnvironment(s.scope)
	if !interactive {
		if environmentID == "" {
			return errors.FailedPreconditionErrorf("no environment selected, use --environment or -E to select an environment")
		}

		return nil
	}

	if environmentID == "" {
		use, err := s.prompter.Confirm("You have not selected an environment. Would you like to select one now?", true)
		if err != nil {
			return err
		}

		if !use {
			return errors.FailedPreconditionErrorf("Please select an environment or use the --environment flag")
		}

		environmentID, err = s.promptForEnvironment(ctx)
		if err != nil {
			return err
		}
		s.scope.GetCurrentContext().EnvironmentID = environmentID
		if err := s.scope.GetCfg().Save(); err != nil {
			return err
		}
		fmt.Println("Changed environment successfully!")
	}

	res, err := s.rig.Environment().List(ctx, &connect.Request[environment.ListRequest]{})
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
		use, err := s.prompter.Confirm(
			"Your selected environment is not available. Would you like to select a new one?", true)
		if err != nil {
			return err
		}

		if !use {
			return errors.FailedPreconditionErrorf("Select an available environment")
		}

		environmentID, err = s.promptForEnvironment(ctx)
		if err != nil {
			return err
		}
		s.scope.GetCurrentContext().EnvironmentID = environmentID
		if err := s.scope.GetCfg().Save(); err != nil {
			return err
		}
		fmt.Println("Changed environment successfully!")
	}

	return nil
}

func (s *Service) AuthUser(ctx context.Context, interactive bool) error {
	user := s.scope.GetCurrentContext().GetAuth().UserID
	if !uuid.UUID(user).IsNil() && user != "" {
		return nil
	}
	if !interactive {
		return errors.UnauthenticatedErrorf("Login to continue")
	}

	loginBool, err := s.prompter.Confirm("You are not logged in. Would you like to login now?", true)
	if err != nil {
		return err
	}
	if !loginBool {
		return errors.UnauthenticatedErrorf("Login to continue")
	}
	return s.login(ctx)
}

func (s *Service) AuthProject(ctx context.Context, interactive bool) error {
	projectID := flags.GetProject(s.scope)
	if !interactive {
		if projectID == "" {
			return errors.FailedPreconditionErrorf("no project selected, use --project/-P or RIG_PROJECT= to select a project")
		}

		return nil
	}

	res, err := s.rig.Project().List(ctx, &connect.Request[project.ListRequest]{})
	if err != nil {
		return err
	}

	if len(res.Msg.Projects) == 0 {
		create, err := s.prompter.Confirm("You have no projects. Would you like to create on now?", true)
		if err != nil {
			return err
		}
		if !create {
			return errors.FailedPreconditionErrorf("Create and select a project to continue")
		}

		if err := s.CreateProject(ctx, "", nil); err != nil {
			return err
		}

		projectID = flags.GetProject(s.scope)

		res, err = s.rig.Project().List(ctx, &connect.Request[project.ListRequest]{})
		if err != nil {
			return err
		}
	}

	if projectID == "" || uuid.UUID(projectID).IsNil() {
		use, err := s.prompter.Confirm("You have not selected a project. Would you like to select one now?", true)
		if err != nil {
			return err
		}
		if !use {
			return errors.FailedPreconditionErrorf("Select a project or use the --project flag to continue")
		}

		if err := s.useProject(ctx); err != nil {
			return err
		}
		projectID = flags.GetProject(s.scope)
	}

	found := false
	for _, p := range res.Msg.GetProjects() {
		if p.GetProjectId() == projectID {
			found = true
			break
		}
	}

	if !found {
		use, err := s.prompter.Confirm("Your selected project is not available. Would you like to select a new one?", true)
		if err != nil {
			return err
		}

		if !use {
			return errors.FailedPreconditionErrorf("Select an available project to continue")
		}

		if use {
			err = s.useProject(ctx)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) CreateProject(ctx context.Context, name string, useNewProject *bool) error {
	var err error
	if name == "" {
		name, err = s.prompter.Input("Project ID:", common.ValidateNonEmptyOpt)
		if err != nil {
			return err
		}
	}

	initializers := []*project.Update{
		{
			Field: &project.Update_Name{
				Name: name,
			},
		},
	}

	res, err := s.rig.Project().Create(ctx, &connect.Request[project.CreateRequest]{
		Msg: &project.CreateRequest{
			Initializers: initializers,
			ProjectId:    name,
		},
	})
	if err != nil {
		return err
	}

	p := res.Msg.GetProject()
	fmt.Printf("Successfully created project %s with ID %s \n", name, p.GetProjectId())

	if useNewProject == nil {
		ok, err := s.prompter.Confirm("Would you like to use this project now?", true)
		if err != nil {
			return err
		}
		useNewProject = &ok
	}

	if *useNewProject {
		s.scope.GetCurrentContext().ProjectID = p.GetProjectId()
		if err := s.scope.GetCfg().Save(); err != nil {
			return err
		}

		fmt.Println("Changed project successfully!")
	}

	return nil
}

func (s *Service) login(ctx context.Context) error {
	u, err := s.prompter.Input("Enter Username or Email:", common.ValidateNonEmptyOpt)
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

	pw, err := s.prompter.Password("Enter Password")
	if err != nil {
		return err
	}

	res, err := s.rig.Authentication().Login(ctx, &connect.Request[authentication.LoginRequest]{
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

	s.scope.GetCurrentContext().GetAuth().UserID = uid.String()
	s.scope.GetCurrentContext().GetAuth().AccessToken = res.Msg.GetToken().GetAccessToken()
	s.scope.GetCurrentContext().GetAuth().RefreshToken = res.Msg.GetToken().GetRefreshToken()
	if err := s.scope.GetCfg().Save(); err != nil {
		return err
	}

	fmt.Println("Login successful!")
	return nil
}

func (s *Service) useProject(ctx context.Context) error {
	var projectID string
	var err error
	listRes, err := s.rig.Project().List(ctx, &connect.Request[project.ListRequest]{})
	if err != nil {
		return err
	}

	var ps []string
	for _, p := range listRes.Msg.GetProjects() {
		ps = append(ps, p.GetName())
	}

	i, _, err := s.prompter.Select("Project: ", ps)
	if err != nil {
		return err
	}

	projectID = listRes.Msg.GetProjects()[i].GetProjectId()

	s.scope.GetCurrentContext().ProjectID = projectID
	if err := s.scope.GetCfg().Save(); err != nil {
		return err
	}

	fmt.Println("Changed project successfully!")

	return nil
}

func (s *Service) promptForEnvironment(ctx context.Context) (string, error) {
	res, err := s.rig.Environment().List(ctx, &connect.Request[environment.ListRequest]{})
	if err != nil {
		return "", err
	}

	var es []string
	for _, e := range res.Msg.GetEnvironments() {
		es = append(es, e.GetEnvironmentId())
	}

	i, _, err := s.prompter.Select("Environment: ", es)
	if err != nil {
		return "", err
	}

	environment := res.Msg.GetEnvironments()[i].GetEnvironmentId()

	return environment, nil
}
