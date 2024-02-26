package base

import (
	"context"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/environment"
	"github.com/rigdev/rig-go-api/api/v1/project"
	"github.com/rigdev/rig-go-sdk"
	"github.com/rigdev/rig/cmd/common"
	"github.com/rigdev/rig/cmd/rig/cmd/cmdconfig"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/scheme"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
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
			fx.Provide(NewKubernetesClient),
			fx.Provide(NewRigClient),
			fx.Provide(func() context.Context { return context.Background() }),

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

func NewKubernetesClient() (client.Client, error) {
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", Flags.KubeConfig)
	if err != nil {
		return nil, err
	}

	if Flags.KubeContext != "" {
		config, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			&clientcmd.ClientConfigLoadingRules{ExplicitPath: Flags.KubeConfig},
			&clientcmd.ConfigOverrides{
				CurrentContext: Flags.KubeContext,
			}).ClientConfig()
		if err != nil {
			return nil, err
		}
	}

	return client.New(config, client.Options{
		Scheme: scheme.New(),
	})
}

func NewRigClient(ctx context.Context) (rig.Client, error) {
	cfg, err := cmdconfig.NewConfig(Flags.RigConfig)
	if err != nil {
		return nil, err
	}

	rc := rig.NewClient(rig.WithSessionManager(&sessionManager{cfg: cfg}))

	// check if we need to authenticate
	projectListResp, err := rc.Project().List(ctx, &connect.Request[project.ListRequest]{})
	if errors.IsUnauthenticated(err) {
		return nil, errors.UnauthenticatedErrorf("You are not authenticated. Please login to continue")
	} else if err != nil {
		return nil, err
	}

	// check if we need to select a project
	found := false
	if Flags.Project != "" {
		for _, p := range projectListResp.Msg.Projects {
			if p.Name == Flags.Project {
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
			projectChoices = append(projectChoices, p.Name)
		}

		_, Flags.Project, err = common.PromptSelect(promptStr, projectChoices)
		if err != nil {
			return nil, err
		}
	}

	envListResp, err := rc.Environment().List(ctx, &connect.Request[environment.ListRequest]{
		Msg: &environment.ListRequest{},
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
		if Flags.Project != "" {
			promptStr = "The environment you selected does not exist. " + promptStr
		}

		environmentChoices := make([]string, 0, len(envListResp.Msg.Environments))
		for _, e := range envListResp.Msg.Environments {
			environmentChoices = append(environmentChoices, e.EnvironmentId)
		}

		_, Flags.Environment, err = common.PromptSelect(promptStr, environmentChoices)
		if err != nil {
			return nil, err
		}
	}

	return rc, nil
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
	s.cfg.Save()
}
