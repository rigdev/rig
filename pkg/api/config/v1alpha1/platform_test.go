package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPlatformConfigValidate(t *testing.T) {
	tests := []struct {
		name        string
		cfg         *PlatformConfig
		expectedErr string
	}{
		{
			name: "invalid github orgRepo",
			cfg: &PlatformConfig{
				Client: Client{
					Git: ClientGit{
						GitHubAuths: []GitHub{{
							OrgRepo: "myorg/myrepo/bad",
						}},
					},
				},
			},
			expectedErr: "client.git[0].orgRepo: Invalid value: \"myorg/myrepo/bad\": `orgRepo` is malformed",
		},
		{
			name: "another invalid github orgRepo",
			cfg: &PlatformConfig{
				Client: Client{
					Git: ClientGit{
						GitHubAuths: []GitHub{{
							OrgRepo: "myorg/",
						}},
					},
				},
			},
			expectedErr: "client.git[0].orgRepo: Invalid value: \"myorg/\": `orgRepo` is malformed",
		},
		{
			name: "good github auth",
			cfg: &PlatformConfig{
				Client: Client{
					Git: ClientGit{
						GitHubAuths: []GitHub{{
							OrgRepo: "myorg/myrepo",
						}},
					},
				},
			},
		},
		{
			name: "another good github auth",
			cfg: &PlatformConfig{
				Client: Client{
					Git: ClientGit{
						GitHubAuths: []GitHub{{
							OrgRepo: "myorg",
						}},
					},
				},
			},
		},
		{
			name: "bad gitlab auth",
			cfg: &PlatformConfig{
				Client: Client{
					Git: ClientGit{
						GitLabAuths: []GitLab{{
							GroupsProject: "group/subgroup1//subgroup2",
						}},
					},
				},
			},
			expectedErr: "client.git[0]: Invalid value: \"group/subgroup1//subgroup2\": `groupsProject` is malformed",
		},
		{
			name: "another bad gitlab auth",
			cfg: &PlatformConfig{
				Client: Client{
					Git: ClientGit{
						GitLabAuths: []GitLab{{
							GroupsProject: "group/subgro:up1/subgroup2",
						}},
					},
				},
			},
			expectedErr: "client.git[0]: Invalid value: \"group/subgro:up1/subgroup2\": `groupsProject` is malformed",
		},
		{
			name: "good gitlab auths",
			cfg: &PlatformConfig{
				Client: Client{
					Git: ClientGit{
						GitLabAuths: []GitLab{
							{GroupsProject: "group"},
							{GroupsProject: "group/subgroup1/subgroup2"},
							{GroupsProject: "group/subgroup1/subgroup2:project"},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.Equal(t, tt.expectedErr, err.Error())
			}
		})
	}
}
