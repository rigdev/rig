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
	"github.com/rigdev/rig/cmd/rig/cmd/cmdconfig"
	"github.com/rigdev/rig/cmd/rig/cmd/flags"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/uuid"
	"github.com/spf13/cobra"
)

type Service struct {
	rig rig.Client
	cfg *cmdconfig.Config
}

func NewService(rig rig.Client, cfg *cmdconfig.Config) *Service {
	return &Service{
		rig: rig,
		cfg: cfg,
	}
}

var (
	OmitUser        = "OMIT_USER"
	OmitProject     = "OMIT_PROJECT"
	OmitEnvironment = "OMIT_ENVIRONMENT"
	OmitCapsule     = "OMIT_CAPSULE"
)

func (s *Service) CheckAuth(ctx context.Context, cmd *cobra.Command, interactive, basicAuth bool) error {
	annotations := common.GetAllAnnotations(cmd)

	var funcs []func(context.Context, bool) error
	if _, ok := annotations[OmitUser]; !ok && !basicAuth {
		funcs = append(funcs, s.authUser)
	}
	if _, ok := annotations[OmitProject]; !ok {
		funcs = append(funcs, s.authProject)
	}
	if _, ok := annotations[OmitEnvironment]; !ok {
		funcs = append(funcs, s.authEnvironment)
	}

	for {
		var retry bool
		var err error
		for _, f := range funcs {
			retry, err = s.handleAuthError(f(ctx, interactive), interactive)
			if err != nil {
				return err
			}
			if retry {
				break
			}
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

	cmdContext := s.cfg.GetCurrentContext()
	str := fmt.Sprintf(
		"There seems to be an issue with the authentication information stored in your current context '%s'",
		cmdContext.Name,
	)
	if !interactive {
		return false, fmt.Errorf("%s: %s", str, origErr)
	}
	fmt.Println(str)
	ok, err := common.PromptConfirm("Do you wish to rebuild this context before proceeding?", true)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, origErr
	}

	s.cfg.DeleteContext(cmdContext.Name)
	if err := s.cfg.CreateContext(cmdContext.Name, cmdContext.GetService().Server, interactive); err != nil {
		return false, err
	}

	return true, nil
}

func (s *Service) authEnvironment(ctx context.Context, interactive bool) error {
	environmentID := flags.GetEnvironment(s.cfg)
	if !interactive {
		if environmentID == "" {
			return errors.FailedPreconditionErrorf("no environment selected, use --environment or -E to select an environment")
		}

		return nil
	}

	if environmentID == "" {
		use, err := common.PromptConfirm("You have not selected an environment. Would you like to select one now?", true)
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
		s.cfg.GetCurrentContext().EnvironmentID = environmentID
		if err := s.cfg.Save(); err != nil {
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
		use, err := common.PromptConfirm(
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
		s.cfg.GetCurrentContext().EnvironmentID = environmentID
		if err := s.cfg.Save(); err != nil {
			return err
		}
		fmt.Println("Changed environment successfully!")
	}

	return nil
}

func (s *Service) authUser(ctx context.Context, interactive bool) error {
	user := s.cfg.GetCurrentAuth().UserID
	if !uuid.UUID(user).IsNil() && user != "" {
		return nil
	}
	if !interactive {
		return errors.UnauthenticatedErrorf("Login to continue")
	}

	loginBool, err := common.PromptConfirm("You are not logged in. Would you like to login now?", true)
	if err != nil {
		return err
	}
	if !loginBool {
		return errors.UnauthenticatedErrorf("Login to continue")
	}
	return s.login(ctx)
}

func (s *Service) authProject(ctx context.Context, interactive bool) error {
	projectID := flags.GetProject(s.cfg)
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
		create, err := common.PromptConfirm("You have no projects. Would you like to create on now?", true)
		if err != nil {
			return err
		}
		if !create {
			return errors.FailedPreconditionErrorf("Create and select a project to continue")
		}

		if err := s.CreateProject(ctx, "", nil); err != nil {
			return err
		}

		projectID = flags.GetProject(s.cfg)

		res, err = s.rig.Project().List(ctx, &connect.Request[project.ListRequest]{})
		if err != nil {
			return err
		}
	}

	if projectID == "" || uuid.UUID(projectID).IsNil() {
		use, err := common.PromptConfirm("You have not selected a project. Would you like to select one now?", true)
		if err != nil {
			return err
		}
		if !use {
			return errors.FailedPreconditionErrorf("Select a project or use the --project flag to continue")
		}

		if err := s.useProject(ctx); err != nil {
			return err
		}
		projectID = flags.GetProject(s.cfg)
	}

	found := false
	for _, p := range res.Msg.GetProjects() {
		if p.GetProjectId() == projectID {
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
		name, err = common.PromptInput("Project ID:", common.ValidateNonEmptyOpt)
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
		ok, err := common.PromptConfirm("Would you like to use this project now?", true)
		if err != nil {
			return err
		}
		useNewProject = &ok
	}

	if *useNewProject {
		s.cfg.GetCurrentContext().ProjectID = p.GetProjectId()
		if err := s.cfg.Save(); err != nil {
			return err
		}

		fmt.Println("Changed project successfully!")
	}

	return nil
}

func (s *Service) login(ctx context.Context) error {
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

	s.cfg.GetCurrentAuth().UserID = uid.String()
	s.cfg.GetCurrentAuth().AccessToken = res.Msg.GetToken().GetAccessToken()
	s.cfg.GetCurrentAuth().RefreshToken = res.Msg.GetToken().GetRefreshToken()
	if err := s.cfg.Save(); err != nil {
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

	i, _, err := common.PromptSelect("Project: ", ps)
	if err != nil {
		return err
	}

	projectID = listRes.Msg.GetProjects()[i].GetProjectId()

	s.cfg.GetCurrentContext().ProjectID = projectID
	if err := s.cfg.Save(); err != nil {
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

	i, _, err := common.PromptSelect("Environment: ", es)
	if err != nil {
		return "", err
	}

	environment := res.Msg.GetEnvironments()[i].GetEnvironmentId()

	return environment, nil
}
