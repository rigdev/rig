package testing

import (
	"testing"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/authentication"
	"github.com/rigdev/rig-go-api/api/v1/environment"
	"github.com/rigdev/rig-go-api/api/v1/project"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig/cmd/rig/cmd"
	"github.com/rigdev/rig/cmd/rig/cmd/cmdconfig"
	authmock "github.com/rigdev/rig/gen/uncommittedmocks/github.com/rigdev/rig-go-api/api/v1/authentication/authenticationconnect"
	environmentmock "github.com/rigdev/rig/gen/uncommittedmocks/github.com/rigdev/rig-go-api/api/v1/environment/environmentconnect"
	projectmock "github.com/rigdev/rig/gen/uncommittedmocks/github.com/rigdev/rig-go-api/api/v1/project/projectconnect"
	rigmock "github.com/rigdev/rig/gen/uncommittedmocks/github.com/rigdev/rig-go-sdk"
	commonmock "github.com/rigdev/rig/gen/uncommittedmocks/github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/pkg/cli"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/dig"
)

var uuid = "7ee202cb-d4be-4bd1-bc8b-a6cd60576567"

type promptMock struct {
	p *commonmock.MockPrompter
}

func newPromptMock(t *testing.T) promptMock {
	return promptMock{
		p: commonmock.NewMockPrompter(t),
	}
}

func (p promptMock) input(value string, numOpts int) {
	var args []any
	for i := 0; i < numOpts; i++ {
		args = append(args, mock.Anything)
	}
	p.p.EXPECT().Input(mock.Anything, args...).Return(value, nil).Once()
}

func (p promptMock) confirm(value bool) {
	p.p.EXPECT().Confirm(mock.Anything, mock.Anything).Return(value, nil).Once()
}

func (p promptMock) password(value string) {
	p.p.EXPECT().Password(mock.Anything).Return(value, nil).Once()
}

func (p promptMock) selectt(idx int, value string) {
	p.p.EXPECT().Select(mock.Anything, mock.Anything).Return(idx, value, nil).Once()
}

type rigMock struct {
	r    *rigmock.MockClient
	auth *authmock.MockServiceClient
	proj *projectmock.MockServiceClient
	env  *environmentmock.MockServiceClient
	t    *testing.T
}

func newRigMock(t *testing.T) *rigMock {
	return &rigMock{
		r: rigmock.NewMockClient(t),
		t: t,
	}
}

func (r *rigMock) Auth() *authmock.MockServiceClient {
	if r.auth == nil {
		r.auth = authmock.NewMockServiceClient(r.t)
		r.r.EXPECT().Authentication().Return(r.auth)
	}
	return r.auth
}

func (r *rigMock) Project() *projectmock.MockServiceClient {
	if r.proj == nil {
		r.proj = projectmock.NewMockServiceClient(r.t)
		r.r.EXPECT().Project().Return(r.proj)
	}
	return r.proj
}

func (r *rigMock) Env() *environmentmock.MockServiceClient {
	if r.env == nil {
		r.env = environmentmock.NewMockServiceClient(r.t)
		r.r.EXPECT().Environment().Return(r.env)
	}
	return r.env
}

type testSuite struct {
	suite.Suite
	rig    *rigMock
	prompt promptMock
	fs     afero.Fs
}

func (s *testSuite) SetupTest() {
	s.T().Setenv("XDG_CONFIG_HOME", "/")
	s.rig = newRigMock(s.T())
	s.prompt = newPromptMock(s.T())
	s.fs = afero.NewMemMapFs()
}

func TestSuite(t *testing.T) {
	suite.Run(t, &testSuite{})
}

func (s *testSuite) run(isInteractive bool, args []string) error {
	module := cli.MakeTestModule(cli.TestModuleInput{
		RigClient:     s.rig.r,
		IsInteractive: isInteractive,
		Prompter:      s.prompt.p,
		FS:            s.fs,
	})
	c := cli.NewSetupContext(module, args)
	c.AddTestCommand = true
	return dig.RootCause(cmd.Run(c))
}

func (s *testSuite) expectProjList(projs ...*project.Project) {
	s.rig.Project().EXPECT().List(mock.Anything, mock.Anything).Return(connect.NewResponse(&project.ListResponse{
		Projects: projs,
	}), nil)
}

func (s *testSuite) expectEnvList(envs ...*environment.Environment) {
	s.rig.Env().EXPECT().List(mock.Anything, mock.Anything).Return(connect.NewResponse(&environment.ListResponse{
		Environments: envs,
	}), nil)
}

func (s *testSuite) expectLoginMail(email, password string, success bool) {
	var resp *connect.Response[authentication.LoginResponse]
	var err error
	if success {
		resp = connect.NewResponse(&authentication.LoginResponse{
			Token: &authentication.Token{
				AccessToken:  "access_token",
				RefreshToken: "refresh_token",
			},
			UserId:   uuid,
			UserInfo: &model.UserInfo{},
		})
	} else {
		err = errors.UnauthenticatedErrorf("oof bad login")
	}

	s.rig.Auth().EXPECT().Login(mock.Anything, connect.NewRequest(&authentication.LoginRequest{
		Method: &authentication.LoginRequest_UserPassword{
			UserPassword: &authentication.UserPassword{
				Identifier: &model.UserIdentifier{
					Identifier: &model.UserIdentifier_Email{
						Email: email,
					},
				},
				Password: password,
			},
		},
	})).Return(resp, err)
}

func (s *testSuite) expectLoginCredentials(clientID, clientSecret string, success bool) {
	var resp *connect.Response[authentication.LoginResponse]
	var err error
	if success {
		resp = connect.NewResponse(&authentication.LoginResponse{
			Token: &authentication.Token{
				AccessToken:  "access_token",
				RefreshToken: "refresh_token",
			},
			UserId:   uuid,
			UserInfo: &model.UserInfo{},
		})
	} else {
		err = errors.UnauthenticatedErrorf("oof bad login")
	}

	s.rig.Auth().EXPECT().Login(mock.Anything, connect.NewRequest(&authentication.LoginRequest{
		Method: &authentication.LoginRequest_ClientCredentials{
			ClientCredentials: &authentication.ClientCredentials{
				ClientId:     clientID,
				ClientSecret: clientSecret,
			},
		},
	})).Return(resp, err)
}

func (s *testSuite) saveConfig(cfg *cmdconfig.Config) {
	cfg2, err := cmdconfig.NewEmptyConfig(s.fs, s.prompt.p)
	s.Require().NoError(err)
	cfg2.Contexts = cfg.Contexts
	cfg2.Services = cfg.Services
	cfg2.Users = cfg.Users
	cfg2.CurrentContextName = cfg.CurrentContextName
	s.Require().NoError(cfg2.Save())
}

func (s *testSuite) getConfig() *cmdconfig.Config {
	cfg, err := cmdconfig.NewConfig("", s.fs, s.prompt.p)
	s.Require().NoError(err)
	return cfg
}

func (s *testSuite) cfgEqual(cfg *cmdconfig.Config) {
	cfg2, err := cmdconfig.NewEmptyConfig(s.fs, s.prompt.p)
	s.Require().NoError(err)
	cfg2.Contexts = cfg.Contexts
	cfg2.CurrentContextName = cfg.CurrentContextName
	cfg2.Services = cfg.Services
	cfg2.Users = cfg.Users
	s.Assert().Equal(cfg2, s.getConfig())
}

func newProject(name string) *project.Project {
	return &project.Project{
		ProjectId: name,
		Name:      name,
	}
}

func newEnv(name string) *environment.Environment {
	return &environment.Environment{
		EnvironmentId: name,
	}
}

func (s *testSuite) Test_empty_config_omit_all() {
	s.Require().NoError(s.run(true, []string{"noop", "cmd1"}))
}

func (s *testSuite) Test_empty_config_omit_none() {
	s.expectLoginMail("mail@example.com", "test123!", true)
	s.expectProjList(newProject("hej"))
	s.expectEnvList(newEnv("prod"))

	s.prompt.input("name", 3)
	s.prompt.input("http://example.com:4747", 2)
	s.prompt.confirm(true) // activate context

	s.prompt.confirm(true)                //  login
	s.prompt.input("mail@example.com", 1) // username
	s.prompt.password("test123!")         // password

	s.prompt.confirm(true) // select project
	s.prompt.selectt(0, "hej")

	s.prompt.confirm(true) // select env
	s.prompt.selectt(0, "prod")

	s.Require().NoError(s.run(true, []string{"noop", "cmd2"}))
}

func (s *testSuite) Test_empty_config_non_interactive() {
	s.Require().NoError(s.run(false, []string{"noop", "cmd1"}))
	s.Require().Error(s.run(false, []string{"noop", "cmd2"}))
}

func (s *testSuite) Test_has_context_but_none_chosen() {
	s.saveConfig(&cmdconfig.Config{
		Contexts: []*cmdconfig.Context{{
			Name:          "ctx",
			ServiceName:   "ctx",
			ProjectID:     "project",
			EnvironmentID: "prod",
		}},
		Services: []*cmdconfig.Service{{
			Name:   "ctx",
			Server: "some-path",
		}},
		Users: []*cmdconfig.User{{
			Name: "ctx",
			Auth: &cmdconfig.Auth{
				UserID:       "user_id",
				AccessToken:  "access_token",
				RefreshToken: "refresh_token",
			},
		}},
	})
	s.prompt.selectt(0, "ctx") // Select config
	s.expectEnvList(newEnv("prod"))
	s.expectProjList(newProject("project"))
	s.Require().NoError(s.run(true, []string{"noop", "cmd2"}))
}

func (s *testSuite) Test_has_full_context() {
	s.saveConfig(&cmdconfig.Config{
		Contexts: []*cmdconfig.Context{{
			Name:          "ctx",
			ServiceName:   "ctx",
			ProjectID:     "project",
			EnvironmentID: "prod",
		}},
		Services: []*cmdconfig.Service{{
			Name:   "ctx",
			Server: "some-path",
		}},
		Users: []*cmdconfig.User{{
			Name: "ctx",
			Auth: &cmdconfig.Auth{
				UserID:       uuid,
				AccessToken:  "access-token",
				RefreshToken: "refresh-token",
			},
		}},
		CurrentContextName: "ctx",
	})
	s.expectProjList(newProject("project"))
	s.expectEnvList(newEnv("prod"))
	s.Require().NoError(s.run(true, []string{"noop", "cmd2"}))
}

func (s *testSuite) Test_help_completion_dont_prompt() {
	s.Require().NoError(s.run(true, []string{"help"}))
	s.Require().NoError(s.run(true, []string{"completion", "bash"}))
}

var _emptyConfig = &cmdconfig.Config{
	Contexts: []*cmdconfig.Context{},
	Services: []*cmdconfig.Service{},
	Users:    []*cmdconfig.User{},
}

func (s *testSuite) Test_auth_activateServiceAccount_no_config_no_host() {
	s.T().Setenv("RIG_CLIENT_ID", "client_id")
	s.T().Setenv("RIG_CLIENT_SECRET", "client_secret")

	s.Require().EqualError(s.run(false, []string{"auth", "activate-service-account"}),
		"no host provided, use `--host` or `RIG_HOST` to specify the host of the Rig platform")

	s.cfgEqual(_emptyConfig)
}

func (s *testSuite) Test_auth_activateServiceAccount_no_config_invalid_host() {
	s.T().Setenv("RIG_HOST", "//example.com")
	s.T().Setenv("RIG_CLIENT_ID", "client_id")
	s.T().Setenv("RIG_CLIENT_SECRET", "client_secret")

	s.Require().EqualError(s.run(false, []string{"auth", "activate-service-account"}),
		"invalid_argument: invalid host, must start with `https://` or `http://`")

	s.cfgEqual(_emptyConfig)
}

func (s *testSuite) Test_auth_activateServiceAccount_no_config() {
	s.T().Setenv("RIG_HOST", "http://example.com:4747")
	s.T().Setenv("RIG_CLIENT_ID", "client_id")
	s.T().Setenv("RIG_CLIENT_SECRET", "client_secret")
	s.expectLoginCredentials("client_id", "client_secret", true)

	s.Require().NoError(s.run(false, []string{"auth", "activate-service-account"}))

	s.cfgEqual(&cmdconfig.Config{
		Contexts: []*cmdconfig.Context{{
			Name:        "service-account",
			ServiceName: "service-account",
		}},
		Services: []*cmdconfig.Service{{
			Name:   "service-account",
			Server: "http://example.com:4747",
		}},
		Users: []*cmdconfig.User{{
			Name: "service-account",
			Auth: &cmdconfig.Auth{
				UserID:       "client_id",
				AccessToken:  "access_token",
				RefreshToken: "refresh_token",
			},
		}},
		CurrentContextName: "service-account",
	})
}

func (s *testSuite) Test_no_prompting_with_context_flag() {
	s.saveConfig(&cmdconfig.Config{
		Contexts: []*cmdconfig.Context{{
			Name:          "context_name",
			ServiceName:   "context_name",
			ProjectID:     "project",
			EnvironmentID: "env",
		}},
		Services: []*cmdconfig.Service{{
			Name:   "context_name",
			Server: "http://example.com:4747",
		}},
		Users: []*cmdconfig.User{{
			Name: "context_name",
			Auth: &cmdconfig.Auth{
				UserID:       "client_id",
				AccessToken:  "access_token",
				RefreshToken: "refresh_token",
			},
		}},
	})
	s.expectProjList(newProject("project"))
	s.expectEnvList(newEnv("env"))
	s.Require().NoError(s.run(true, []string{"noop", "cmd2", "--context", "context_name"}))
}

func (s *testSuite) Test_no_prompting_config_commands() {
	s.Require().NoError(s.run(true, []string{"config", "view"}))
	s.Require().NoError(s.run(true, []string{"config", "list-contexts"}))
}

func (s *testSuite) Test_config_init() {
	// Create context 1
	s.prompt.input("context1", 3)
	s.prompt.input("http://example.com:4747", 2)
	s.prompt.confirm(true)
	s.Require().NoError(s.run(true, []string{"config", "init"}))

	// Create context 2
	s.prompt.input("context2", 3)
	s.prompt.input("http://example.com:4748", 2)
	s.prompt.confirm(true)
	s.Require().NoError(s.run(true, []string{"config", "init"}))
}
