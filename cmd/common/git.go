package common

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	"github.com/rigdev/rig-go-api/api/v1/environment"
	"github.com/rigdev/rig-go-api/model"
	"github.com/rigdev/rig-go-sdk"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/proto"
)

type GitFlags struct {
	Repository     string
	Branch         string
	CapsulePath    string
	CommitTemplate string
	Environments   string
	Disable        bool
	Enable         bool
}

var (
	capsulePathDefault    = "{{ .Project }}/{{ .Capsule }}/{{ .Environment}}.yaml"
	commitTemplateDefault = "Updating {{ .Type }} {{ .Name }} on behalf of {{ .Author }}"
)

func (g *GitFlags) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&g.Repository, "repository", "", "URL to the git repository to use.")
	cmd.Flags().StringVar(&g.Branch, "branch", "", "The branch of the git repository to use.")
	cmd.Flags().StringVar(&g.CapsulePath, "capsule-path",
		capsulePathDefault,
		"The templated path to the file containing the capsule spec.",
	)
	cmd.Flags().StringVar(&g.CommitTemplate, "commit-template",
		commitTemplateDefault, "The commit template when Rig creates commits on behalf of a user.",
	)
	cmd.Flags().StringVar(&g.Environments, "environments", "",
		"The environment filter to use. Can be one of 'all', 'all+ephem' or a comma separated list of env names.",
	)
	cmd.Flags().BoolVar(&g.Disable, "disable", false, "disable git store")
}

func (g *GitFlags) FeedStore(store *model.GitStore) {
	updated := false
	if g.Repository != "" {
		store.Repository = g.Repository
		updated = true
	}
	if g.Branch != "" {
		store.Branch = g.Branch
		updated = true
	}
	if g.CapsulePath != "" {
		if g.CapsulePath != capsulePathDefault || store.CapsulePath == "" {
			store.CapsulePath = g.CapsulePath
			updated = true
		}
	}
	if g.CommitTemplate != "" {
		if g.CommitTemplate != commitTemplateDefault || store.CommitTemplate == "" {
			store.CommitTemplate = g.CommitTemplate
			updated = true
		}
	}
	if g.Environments != "" {
		store.Environments = ParseEnvironmentFilter(g.Environments)
		updated = true
	}
	if g.Disable {
		store.Disabled = g.Disable
	} else if updated {
		store.Disabled = false
	}
}

func UpdateGit(
	ctx context.Context, rig rig.Client, flags GitFlags, isInteractive bool, prompter Prompter, gitStore *model.GitStore,
) (*model.GitStore, error) {
	if gitStore == nil {
		gitStore = &model.GitStore{}
	}
	gitStore = proto.Clone(gitStore).(*model.GitStore)
	flags.FeedStore(gitStore)
	var missing string
	if gitStore.GetRepository() == "" {
		missing = "--repository"
	} else if gitStore.GetBranch() == "" {
		missing = "--branch"
	} else if gitStore.GetEnvironments() == nil {
		missing = "--environments"
	}

	if isInteractive {
		if missing != "" {
			return nil, fmt.Errorf("%s must be given", missing)
		}
	} else if missing != "" {
		envResp, err := rig.Environment().List(ctx, connect.NewRequest(&environment.ListRequest{}))
		if err != nil {
			return nil, err
		}
		if gitStore, err = PromptGitStore(prompter, gitStore, envResp.Msg.GetEnvironments()); err != nil {
			return nil, err
		}
	}
	return gitStore, nil
}

func ParseEnvironmentFilter(envString string) *model.EnvironmentFilter {
	switch envString {
	case "all":
		return &model.EnvironmentFilter{
			Filter: &model.EnvironmentFilter_All_{
				All: &model.EnvironmentFilter_All{
					IncludeEphemeral: false,
				},
			},
		}
	case "all+ephem":
		return &model.EnvironmentFilter{
			Filter: &model.EnvironmentFilter_All_{
				All: &model.EnvironmentFilter_All{
					IncludeEphemeral: true,
				},
			},
		}
	default:
		return &model.EnvironmentFilter{
			Filter: &model.EnvironmentFilter_Selected_{
				Selected: &model.EnvironmentFilter_Selected{
					EnvironmentIds: strings.Split(envString, ","),
				},
			},
		}
	}
}

func PromptGitStore(
	prompter Prompter,
	gitStore *model.GitStore,
	envs []*environment.Environment,
) (*model.GitStore, error) {
	if gitStore == nil {
		gitStore = &model.GitStore{}
	}
	gitStore = proto.Clone(gitStore).(*model.GitStore)
	fields := []string{
		"Disabled",
		"Repository",
		"Branch",
		"Capsule Path",
		"Commit Template",
		"Environments",
		"Done",
	}

	for {
		i, _, err := prompter.Select("Select the field to update (CTRL + c to cancel)", fields)
		if err != nil {
			return nil, err
		}

		switch i {
		case 0:
			disable, err := prompter.Confirm("Disable Git store", false)
			if err != nil {
				if ErrIsAborted(err) {
					continue
				}
				return nil, err
			}

			gitStore.Disabled = disable
		case 1:
			repo, err := prompter.Input("Enter the repository URL",
				ValidateNonEmptyOpt, InputDefaultOpt(gitStore.GetRepository()))
			if err != nil {
				if ErrIsAborted(err) {
					continue
				}
				return nil, err
			}

			gitStore.Repository = repo
		case 2:
			branch, err := prompter.Input("Enter the branch",
				ValidateNonEmptyOpt,
				InputDefaultOpt(gitStore.GetBranch()),
			)
			if err != nil {
				if ErrIsAborted(err) {
					continue
				}
				return nil, err
			}

			gitStore.Branch = branch
		case 3:
			path, err := prompter.Input("Enter the capsule path",
				ValidateNonEmptyOpt,
				InputDefaultOpt(StringOr(gitStore.GetCapsulePath(), "{{ .Project }}/{{ .Capsule }}/{{ .Environment}}.yaml")),
			)
			if err != nil {
				if ErrIsAborted(err) {
					continue
				}
				return nil, err
			}

			gitStore.CapsulePath = path
		case 4:
			template, err := prompter.Input("Enter the commit template",
				ValidateNonEmptyOpt,
				InputDefaultOpt(
					StringOr(gitStore.GetCommitTemplate(), "Updating {{ .Type }} {{ .Name }} on behalf of {{ .Author }}"),
				),
			)
			if err != nil {
				if ErrIsAborted(err) {
					continue
				}
				return nil, err
			}

			gitStore.CommitTemplate = template
		case 5:
			if gitStore.Environments == nil {
				gitStore.Environments = &model.EnvironmentFilter{}
			}
			gitStore.Environments, err = PromptEnvironmentFilter(prompter, gitStore.GetEnvironments(), envs)
			if err != nil {
				if ErrIsAborted(err) {
					continue
				}
				return nil, err
			}
		default:
			return gitStore, nil
		}
	}
}

func PromptEnvironmentFilter(
	prompter Prompter,
	filter *model.EnvironmentFilter,
	environments []*environment.Environment,
) (*model.EnvironmentFilter, error) {
	if filter == nil {
		filter = &model.EnvironmentFilter{}
	}
	filter = proto.Clone(filter).(*model.EnvironmentFilter)

	for {
		var envs []string
		for _, e := range environments {
			env := e.GetEnvironmentId()
			if slices.Contains(filter.GetSelected().GetEnvironmentIds(), e.GetEnvironmentId()) {
				env += " *"
			}

			envs = append(envs, env)
		}

		all := "All"
		allEphemeral := "All + Ephemeral"
		if filter.GetAll() != nil {
			if filter.GetAll().GetIncludeEphemeral() {
				allEphemeral += " *"
			} else {
				all += " *"
			}
		}

		envs = append(envs, all, allEphemeral, "Done")

		i, _, err := prompter.Select("Select Environments (select current environments marked by * to remove)", envs)
		if err != nil {
			return nil, err
		}

		if i == len(envs)-1 {
			break
		}

		if i == len(envs)-2 {
			if filter.GetAll() == nil {
				filter.Filter = &model.EnvironmentFilter_All_{
					All: &model.EnvironmentFilter_All{},
				}
			}

			filter.GetAll().IncludeEphemeral = true
		} else if i == len(envs)-3 {
			if filter.GetAll() == nil {
				filter.Filter = &model.EnvironmentFilter_All_{
					All: &model.EnvironmentFilter_All{},
				}
			}

			filter.GetAll().IncludeEphemeral = false
		} else {
			env := environments[i]

			if filter.GetSelected() == nil {
				filter.Filter = &model.EnvironmentFilter_Selected_{
					Selected: &model.EnvironmentFilter_Selected{},
				}
			}

			if i := slices.Index(filter.GetSelected().GetEnvironmentIds(), env.GetEnvironmentId()); i != -1 {
				filter.GetSelected().EnvironmentIds = slices.Delete(filter.GetSelected().GetEnvironmentIds(), i, i+1)
			} else {
				filter.GetSelected().EnvironmentIds = append(filter.GetSelected().GetEnvironmentIds(), env.GetEnvironmentId())
			}
		}

	}

	return filter, nil
}
