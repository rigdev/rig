package base

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/authentication"
	"github.com/rigdev/rig-go-api/api/v1/environment"
	"github.com/rigdev/rig-go-api/api/v1/project"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/cmdconfig"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/roclient"
	"github.com/rigdev/rig/pkg/scheme"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var options []fx.Option

func Register(f interface{}) func(cmd *cobra.Command, args []string) error {
	options = append(options,
		fx.Provide(f),
	)

	return func(cmd *cobra.Command, args []string) error {
		var opts []fx.Option
		f := fx.New(
			fx.Supply(cmd),
			fx.Supply(args),
			fx.Provide(NewKubernetesClient, NewKubernetesReader),
			fx.Provide(NewRigClient, NewOperatorClient),
			fx.Provide(func() context.Context { return context.Background() }),
			fx.Provide(scheme.New),

			fx.Invoke(f),

			fx.NopLogger,
			fx.Options(opts...),
		)

		if err := f.Start(context.Background()); err != nil {
			return err
		}
		if err := f.Stop(context.Background()); err != nil {
			return err
		}
		return f.Err()
	}
}

func GetRestConfig() (*rest.Config, error) {
	config, err := clientcmd.BuildConfigFromFlags("", Flags.KubeConfig)
	if err != nil {
		return config, err
	}

	if Flags.KubeContext != "" {
		config, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			&clientcmd.ClientConfigLoadingRules{ExplicitPath: Flags.KubeConfig},
			&clientcmd.ConfigOverrides{
				CurrentContext: Flags.KubeContext,
			}).ClientConfig()
		if err != nil {
			return config, err
		}
	}

	return config, nil
}

func NewKubernetesClient() (client.Client, *rest.Config, error) {
	// use the current context in kubeconfig
	config, err := GetRestConfig()
	if err != nil {
		return nil, config, err
	}

	cc, err := client.New(config, client.Options{
		Scheme: scheme.New(),
	})
	return cc, config, err
}

func NewKubernetesReader(cc client.Client) (client.Reader, error) {
	if Flags.KubeFile != "" {
		return roclient.NewReaderFromFile(Flags.KubeFile, cc.Scheme())
	}

	return cc, nil
}

func NewRigClient(ctx context.Context, fs afero.Fs, prompter common.Prompter) (rig.Client, error) {
	cfg, err := cmdconfig.NewConfig(Flags.RigConfig, fs, prompter)
	if err != nil {
		return nil, err
	}

	if cfg.GetCurrentContext() == nil {
		// TODO Catch with FX instead?
		return nil, fmt.Errorf("no rig context. Run `rig config init` to setup one up")
	}

	sessionManager := &sessionManager{cfg: cfg}

	host := cfg.GetCurrentService().Server
	if Flags.RigContext != "" {
		serv, err := cfg.GetService(Flags.RigContext)
		if err != nil {
			return nil, err
		}

		host = serv.Server
	}

	rc := rig.NewClient(rig.WithSessionManager(sessionManager), rig.WithHost(host))

	// check if we need to authenticate
	projectListResp, err := rc.Project().List(ctx, &connect.Request[project.ListRequest]{})
	if errors.IsUnauthenticated(err) {
		tokens, err := rigLogin(ctx, rc, prompter)
		if err != nil {
			return nil, err
		}
		sessionManager.SetAccessToken(tokens.AccessToken, tokens.RefreshToken)
		projectListResp, err = rc.Project().List(ctx, &connect.Request[project.ListRequest]{})
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	// check if we need to select a project
	found := false
	if Flags.Project != "" {
		for _, p := range projectListResp.Msg.Projects {
			if p.GetProjectId() == Flags.Project {
				found = true
				break
			}
		}
	}

	if !found {
		promptStr := "Select a project to continue"
		if Flags.Project != "" {
			promptStr = "The project you selected does not exist. " + promptStr
		}

		projectChoices := make([]string, 0, len(projectListResp.Msg.Projects))
		for _, p := range projectListResp.Msg.Projects {
			projectChoices = append(projectChoices, p.GetProjectId())
		}

		_, Flags.Project, err = prompter.Select(promptStr, projectChoices)
		if err != nil {
			return nil, err
		}
	}

	envListResp, err := rc.Environment().List(ctx, &connect.Request[environment.ListRequest]{
		Msg: &environment.ListRequest{
			ProjectFilter: Flags.Project,
		},
	})
	if err != nil {
		return nil, err
	}

	// check if we need to select an environment
	found = false
	if Flags.Environment != "" {
		for _, e := range envListResp.Msg.Environments {
			if e.EnvironmentId == Flags.Environment {
				found = true
				break
			}
		}
	}

	if !found {
		promptStr := "Select an environment to continue"
		if Flags.Environment != "" {
			promptStr = "The environment you selected does not exist, or is not active for the chosen project." + promptStr
		}

		environmentChoices := make([]string, 0, len(envListResp.Msg.Environments))
		for _, e := range envListResp.Msg.Environments {
			environmentChoices = append(environmentChoices, e.EnvironmentId)
		}

		_, Flags.Environment, err = prompter.Select(promptStr, environmentChoices)
		if err != nil {
			return nil, err
		}
	}

	return rc, nil
}

func rigLogin(ctx context.Context, rc rig.Client, prompter common.Prompter) (*authentication.Token, error) {
	email, err := prompter.Input(
		"You are not logged in on the Rig platform. Please do so to migrate using the platform.\nEmail:",
		common.ValidateEmailOpt,
	)
	if err != nil {
		return nil, err
	}

	password, err := prompter.Password("Password: ")
	if err != nil {
		return nil, err
	}

	loginResp, err := rc.Authentication().Login(ctx, &connect.Request[authentication.LoginRequest]{
		Msg: &authentication.LoginRequest{
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
		},
	})
	if err != nil {
		return nil, err
	}

	return loginResp.Msg.GetToken(), nil
}

type sessionManager struct {
	cfg *cmdconfig.Config
}

func (s *sessionManager) GetAccessToken() string {
	if Flags.RigContext != "" {
		for _, u := range s.cfg.Users {
			if u.Name == Flags.RigContext {
				return u.Auth.AccessToken
			}
		}
		return ""
	}

	return s.cfg.GetCurrentAuth().AccessToken
}

func (s *sessionManager) GetRefreshToken() string {
	if Flags.RigContext != "" {
		for _, u := range s.cfg.Users {
			if u.Name == Flags.RigContext {
				return u.Auth.AccessToken
			}
		}
		return ""
	}

	return s.cfg.GetCurrentAuth().AccessToken
}

func (s *sessionManager) SetAccessToken(accessToken string, refreshToken string) {
	var auth *cmdconfig.Auth
	if Flags.RigContext != "" {
		for _, u := range s.cfg.Users {
			if u.Name == Flags.RigContext {
				auth = u.Auth
			}
		}
	} else {
		auth = s.cfg.GetCurrentAuth()
	}

	if auth == nil {
		return
	}

	auth.AccessToken = accessToken
	auth.RefreshToken = refreshToken
	if err := s.cfg.Save(); err != nil {
		fmt.Println("Failed to save the new access token")
	}
}
