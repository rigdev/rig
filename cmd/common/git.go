package common

import (
	"github.com/rigdev/rig-go-api/api/v1/environment"
	"github.com/rigdev/rig-go-api/model"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/proto"
)

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
